export GOMAXPROCS=8
#go run  main.go -n 5 -v -url http://eclipse.c3sl.ufpr.br/technology/epp/downloads/release/2018-09/R/eclipse-jee-2018-09-win32-x86_64.zip  -o eclipse.zip && unzip -t eclipse.zip > /dev/null && echo Test OK
go run  main.go  -url http://eclipse.c3sl.ufpr.br/technology/epp/downloads/release/2018-09/R/eclipse-jee-2018-09-win32-x86_64.zip  -o eclipse.zip && unzip -t eclipse.zip > /dev/null && echo Test OK
rm eclipse.zip 2> /dev/null