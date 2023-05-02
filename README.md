# homevision-takehome

This is a script that downloads the images of the first 10 houses provided by the HomeVision API.

To run it, open a terminal and run:
```
go run main.go
```

It will create the directory `tmp` and save the photos inside it, using the following format for the file names:
```
[house_id]-[address].[ext]
```
