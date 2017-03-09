#!/usr/bin/env bash
set -e

bin/getter -template

mkdir -p /opt/thingsweather/
cp bin/getter /opt/thingsweather/

mkdir -p /etc/thingsweather/
cp sampleConfig.json /etc/thingsweather/config.json

getent group thingsweather || groupadd thingsweather
id -u thingsweather &>/dev/nul || useradd -r -s /bin/false -g thingsweather thingsweather

# Create service file
cat >/lib/systemd/system/thingsweather.service <<EOL
[Unit]
Description=Things weather daemon
Requires=network.target
After=rsyslog.service network.target influxd.service

[Service]
Type=notify
PermissionsStartOnly=true
ExecStart=/opt/thingsweather/getter -config=/etc/thingsweather/config.json
RestartSec=5
Restart=always
TimeoutStartSec=30
TimeoutStopSec=10
User=thingsweather
Group=thingsweather
OOMScoreAdjust=100
StandardInput=null
StandardOutput=syslog
StandardError=inherit
SyslogIdentifier=WeatherGetter
SyslogFacility=local0
SyslogLevel=debug
SyslogLevelPrefix=true
Environment=TZ=UTC

[Install]
WantedBy=multi-user.target
EOL

systemctl daemon-reload

printf "Done\n"
printf "Enable ThinsWeather on startup with \"systemctl enable thingsweather.service\"\n"
printf "Start ThinsWeather on startup with \"systemctl start thingsweather.service\"\n"
