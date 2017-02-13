VERSION=0.1.6

all:$(OUT_DIR)/jsonline
$(OUT_DIR)/jsonline:main.go
	gox  \
		-output "bin/{{.Dir}}_${VERSION}/{{.OS}}_{{.Arch}}/{{.Dir}}" \
		-osarch "linux/arm" github.com/FarmRadioHangar/jsonline


tar:
	cd bin/  && tar -zcvf jsonline_${VERSION}.tar.gz jsonline_${VERSION}
