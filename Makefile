J_VERSION=0.1.4
NAME=jsonline_$(J_VERSION)
OUT_DIR=bin/linux_arm/jsonline_$(J_VERSION)

all:$(OUT_DIR)/jsonline bundle
$(OUT_DIR)/jsonline:main.go
	gox  \
		-output "bin/{{.OS}}_{{.Arch}}/{{.Dir}}_$(J_VERSION)/{{.Dir}}" \
		-osarch "linux/arm" github.com/FarmRadioHangar/jsonline

bundle:$(OUT_DIR)/jsonline_telegraf.service $(OUT_DIR)/Makefile

$(OUT_DIR)/jsonline_telegraf.service:scripts/jsonline_telegraf.service
	cp scripts/jsonline_telegraf.service $(OUT_DIR)

$(OUT_DIR)/Makefile:scripts/Makefile
	cp scripts/Makefile $(OUT_DIR)

tar:
	tar -zcvf jsonline_$(J_VERSION).tar.gz  $(OUT_DIR)
