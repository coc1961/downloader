docker run -d --rm --name mycntlm -e CNTLM_PROXY="proxy:8100" -e CNTLM_NO_PROXY=* -p 8888:3128  bachp/cntlm 

#docker logs -f mycntlm &

export http_proxy=http://127.0.0.1:8888

export GOMAXPROCS=8
go run  main.go -n 5 -v -url http://eclipse.c3sl.ufpr.br/technology/epp/downloads/release/2018-09/R/eclipse-jee-2018-09-win32-x86_64.zip  -o eclipse.zip && unzip -t eclipse.zip > /dev/null && echo Test OK
rm eclipse.zip 2> /dev/null


docker stop mycntlm 