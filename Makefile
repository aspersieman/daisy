
all: daisy

fruit.o: fruit.c fruit.h
	gcc -c -g -O0 -Wall fruit.c

daisy.o: main.c fruit.h
	gcc -c -g -O0 -Wall main.c

daisy: fruit.o daisy.o
	gcc -o daisy fruit.o main.o -lyaml

clean:
	rm -f daisy
	rm -f *.o core
