package download

// Download File, permite descargar un archivo pedazos con varios hilos simultaneos
//
// Utilzando la funcion
//
// func DownloadFile(resourceURL *url.URL, workers int64, out *os.File, listener func(status []*progressReader)) {
//
// recibe la url, la cantidad de hilos, el archhivo de salida, y una funcion que recibe el status con el progreso
// de la descarga

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"sync"
	"time"
)

const sleepTipme time.Duration = 500
const secondsTimeout = 10
const nanosecondsTimeout = secondsTimeout * 1000000000

// Estructura de descarga parcial
type partialDownload struct {
	resourceURL *url.URL // Url de descarga
	rangeHeader string   // bytes a descargar
	out         *os.File // archivo de salida
	err         error    // error ?
	chunkSize   int64    // bytes a descargar
	chunkStart  int64    // posicion de inicio
	chunkEnd    int64    // posicion de fin
}

// Creo partialDownload
func createPartialDownload(resourceURL *url.URL, chunkStart int64, chunkEnd int64, out *os.File) *partialDownload {
	rangeHeader := fmt.Sprintf("bytes=%d-%d", chunkStart, chunkEnd)
	return &partialDownload{resourceURL, rangeHeader, out, nil, chunkEnd - chunkStart, chunkStart, chunkEnd}
}

//CreateClient Crea un cliente http con o sin proxy
func CreateClient() *http.Client {
	proxy := os.Getenv("http_proxy")
	var proxyURL *url.URL
	var transport *http.Transport
	if proxy != "" {
		proxyURL, _ = url.Parse(proxy)
	}
	if proxyURL != nil {
		transport = &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		}
	}
	var client http.Client
	if transport != nil {
		client = http.Client{
			Transport: transport,
		}
	}

	//client.Timeout = time.Second * 5
	return &client
}

// Descarga Parcial
func (p *partialDownload) download(progressArray *[]*progressReader, wg *sync.WaitGroup) {
	defer wg.Done()

	var wrapReader *progressReader

	for i := 0; ; i++ {
		// Seteo el rango a descargar
		p.rangeHeader = fmt.Sprintf("bytes=%d-%d", p.chunkStart, p.chunkEnd)

		// Request
		req, error := http.NewRequest("GET", p.resourceURL.String(), nil)
		if error != nil {
			p.err = error
			continue
		}

		// Seteo rango de descarga
		req.Header.Add("Range", p.rangeHeader)
		var client = CreateClient()

		resp, error := client.Do(req)
		if error != nil {
			p.err = error
			continue
		}

		// Archivo de descarga parcial
		out, err := os.OpenFile(p.out.Name(), os.O_RDWR, os.ModePerm)
		if err != nil {
			p.err = err
			return
		}
		p.out = out

		// Me posiciono en la posicion en donde debo descargar
		p.out.Seek(p.chunkStart, os.SEEK_SET)

		// Agrego al array para desplegar el porc de descarga
		if i == 0 {
			// Creo un emboltorio de reader para que setee valores
			// de progreso de descarga para alimentar la barra de %
			wrapReader = createprogressReader(&resp.Body, p.chunkSize)
			*progressArray = append(*progressArray, wrapReader)
		} else {
			// Reproceso
			// Actualizo el progressReader
			wrapReader.len = p.chunkSize
			wrapReader.pos = 0
			wrapReader.reader = &resp.Body
		}

		// Descargo!
		_, error = io.Copy(p.out, wrapReader)
		p.err = error
		if error != nil {
			// Si Hay Error Reproceso
			// Recalculo la posicion
			p.chunkStart += wrapReader.pos
			p.chunkSize = p.chunkEnd - p.chunkStart
			out.Close()
		} else {
			break
		}
	}
	return
}

// Creo el objeto para alimentar la barra de progreso
func createprogressReader(reader *io.ReadCloser, len int64) *progressReader {
	ret := &progressReader{reader, len, 0, time.Now()}
	return ret
}

// Status chunck download status
type Status interface {
	Progress() int64
}

//progressReader  Envoltorio de reader que guarda el % de descarga
type progressReader struct {
	reader   *io.ReadCloser // Reader original
	len      int64          // total a descargar
	pos      int64          // bytes procesados
	lastRead time.Time      // Ultima Lectura
}

//Progress Retorna al porcentaje de la descarga realizada
func (r *progressReader) Progress() int64 {
	return int64(r.pos / (r.len / 100))
}

// Actualizo % de descarga y realizo la lectura real
func (r *progressReader) Read(p []byte) (n int, err error) {
	r.lastRead = time.Now()
	lei, err := (*(r.reader)).Read(p)
	r.pos += int64(lei)
	return lei, err
}

// File descarga un archivo en partes procesadas en procesos simultaneos
func File(resourceURL *url.URL, workers int64, out *os.File, listener func(status []Status)) {

	client := CreateClient()
	res, err := client.Head(resourceURL.String())
	if err != nil {
		log.Fatalf("error requesting HEAD of file: %v", err)
	}

	if res.StatusCode >= 300 {
		log.Fatalf("unexpected status code received from server: %s", res.Status)
	}

	acceptRange := res.Header.Get("accept-ranges")
	if acceptRange != "bytes" && workers > 1 {
		log.Fatalf("remote server does not accept range downloads")
	}

	contentLength, err := strconv.ParseInt(res.Header.Get("content-length"), 10, 64)
	if err != nil || contentLength == 0 {
		log.Fatalf("remote server content-length is invalid")
	}

	// Inicializo Variables
	progressBarArray := make([]*progressReader, 0)
	partialDownloadArray := make([]*partialDownload, 0)

	// Reservo Espacio en el Archivo de Salida
	out.Truncate(contentLength)

	// Calculo el tamaño del chunk tamaño / hilos
	chunkSize := contentLength / workers

	wg := sync.WaitGroup{}
	wg.Add(int(workers))

	// Lanzo el Hilo de Verificacion de Timeout
	go timeoutVerify(&progressBarArray)

	// Creo los downloaders parciales
	for i := int64(0); i < workers; i++ {
		chunkStart, chunkEnd := chunkSize*i, (chunkSize*i)+chunkSize-1
		if i+1 == workers {
			chunkEnd = contentLength
		}

		// creo un object de descarga parcial con
		// url del archivo, byte de inicio de la descarga, byte de fin, archivo de salida
		tmp := createPartialDownload(resourceURL, chunkStart, chunkEnd, out)
		partialDownloadArray = append(partialDownloadArray, tmp)

		go func() {
			// Comienzo Descarga parcial
			tmp.download(&progressBarArray, &wg)
		}()
	}

	// Muestro Barra de Progreso
	if listener != nil {
		go func() {
			for {
				// Convierto en Objetos Status
				statusArray := make([]Status, len(progressBarArray))
				for i, value := range progressBarArray {
					statusArray[i] = value
				}
				listener(statusArray)
				time.Sleep(time.Millisecond * sleepTipme)
			}
		}()
	}

	wg.Wait()

	// Cierro archivo de salida
	out.Close()

	// Cierro archivos parciales
	for _, v := range partialDownloadArray {
		if v.out != nil {
			v.out.Close()
		}
	}
	// Hay Errores ?
	for _, v := range partialDownloadArray {
		if v.err != nil {
			go os.Remove(out.Name())
			log.Fatal(v.err)
		}
	}
	time.Sleep(time.Millisecond * sleepTipme)

}

// Verifico el Timeout para detectar caidas en la conexion
func timeoutVerify(progressBarArray *[]*progressReader) {
	for {
		for _, ret := range *progressBarArray {
			diff := time.Now().Sub(ret.lastRead)
			if diff > nanosecondsTimeout {
				(*(ret.reader)).Close()
				ret.lastRead = time.Now()
			}
		}
		time.Sleep(time.Second)
	}
}
