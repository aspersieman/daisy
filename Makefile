
all: daisy

daisy: daisy.o
	gcc -o daisy daisy.o

daisy.o: daisy.c
	gcc -c -g -O0 -Wall daisy.c

scan.o: scan.c
	gcc -c -g -O0 -Wall scan.c

scan: scan.o
	gcc -o scan scan.o -lyaml
