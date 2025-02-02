
all: daisy

daisy: daisy.o
	gcc -o daisy daisy.o -lyaml

daisy.o: daisy.c daisy.h
	gcc -c -g -O0 -Wall daisy.c

parse.o: parse.c fruit.h
	gcc -c -g -O0 -Wall parse.c

parse: fruit.o parse.o
	gcc -o parse fruit.o parse.o -lyaml

clean:
	rm -f daisy parse
	rm -f *.o core
