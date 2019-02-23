# Dowloader

#### Gestor de Descarga 

Programa para descarga de archivos, permite partir la descarga en varios hilos simultáneos.

### parámetros:

    -url file url (required)
    -o   output file (required)
    -n   number of concurent downloads (Optional) 
    -v   show progress (Optional)

### Ejemplo de Uso

`downloader -n 5 -v -url http://myhost.com/eclipse.zip -o eclipse.zip`

### Notas

Se implementa recuperación de errores en caso de micro cortes en la comunicación
