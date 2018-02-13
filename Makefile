lib = sum.so
test = test

python-libs = `pkg-config --cflags --libs python3`

coap.so: coap.c coap.go
	#gcc -Wall -fPIC -shared -o coap.so $(python-libs)  coap.c
	go build -v -buildmode=c-shared -o coap.so

$(lib): *.go  
	go build -buildmode=c-shared -o $(lib) sum.go

test: coap.c
	gcc -Wall -fPIC -shared -o coap.so $(python-libs)  coap.c

main: *.c $(lib)
	gcc -Wall -o main main.c ./$(lib)

clean:
	rm -rf *.so
	rm -rf *.h
	