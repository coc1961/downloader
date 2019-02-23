# Dowloader

#### Console Gownload Manager

## Install

`go get -u github.com/coc1961/downloader`

## Use

Program for downloading files, allows to split the download into several simultaneous threads.

### Parameters:

    -url file url (required)
    -o   output file (Optional, Recomended)
    -n   number of concurent downloads (Optional) 
    -v   show progress (Optional)

### Example

`downloader -n 5 -v -url http://myhost.com/eclipse.zip -o eclipse.zip`

### Note

Error recovery is implemented in case of micro cuts in communication
