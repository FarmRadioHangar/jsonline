
all: /usr/local/bin/jsonline /etc/systemd/system/jsonline_telegraf.service
/usr/local/bin/jsonline:jsonline
	cp jsonline /usr/local/bin
/etc/systemd/system/jsonline_telegraf.service:jsonline_telegraf.service
	cp jsonline_telegraf.service /etc/systemd/system/


install:
	systemctl enable jsonline_telegraf.service
