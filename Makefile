compile:
	gox  -output "bin/{{.OS}}_{{.Arch}}/{{.Dir}}" github.com/FarmRadioHangar/jsonline
