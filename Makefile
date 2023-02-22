ifndef GO_LIB_PATH
	GO_LIB_PATH=./pferret/lib/
endif

ifndef VERSION
	VERSION=$(shell cat version)
endif

build:
	$(go version)
	cd ${GO_LIB_PATH} && go build -buildmode c-shared -o libferret.so

clean:
	rm -rf
	rm -rf  dist pferret.egg-info dist build


publish-package:
	python setup.py bdist_wheel upload -r ferret
