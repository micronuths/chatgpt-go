rm -rf "$(pwd)/service.file/chatgpt-go.service" /etc/systemd/system/chatgpt-go.service
mkdir -p "$(pwd)/service.file"
echo "\
[Unit]
Description=chatgpt-go
After=network.target
 
[Service]
Type=simple
WorkingDirectory=$(pwd)
ExecStart=$(pwd)/chatgpt-go
Restart=on-failure
 
[Install]
WantedBy=multi-user.target\
" >> "$(pwd)/service.file/chatgpt-go.service"

cp $(pwd)/service.file/chatgpt-go.service /etc/systemd/system/
systemctl daemon-reload
systemctl enable chatgpt-go.service