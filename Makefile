export GOPATH=$(shell pwd)
dep = ${GOPATH}/bin/dep
python-libs = `pkg-config --cflags --libs python3`
srcdir = src/pycoap
src = $(srcdir)/coap.c $(srcdir)/coap.go
vendor = src/pycoap/vendor
target = ${GOPATH}/pycoap.so

$(target): $(dep) $(src) $(vendor)
	#gcc -Wall -fPIC -shared -o coap.so $(python-libs)  coap.c
	cd $(srcdir); go build -v -buildmode=c-shared -o $(target)

$(dep):
	go get -u github.com/golang/dep/cmd/dep

$(vendor):
	cd $(srcdir); dep ensure

clean:
	rm -rf $(srcdir)/*.so
	rm -rf $(srcdir)/*.h
	