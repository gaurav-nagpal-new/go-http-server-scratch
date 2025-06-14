# go-http-server-scratch
A http server in golang from scratch

<!-- Run server in your local-->

# step 1
go mod tidy

# step 2 (Navigate to server directory)
cd app/server

# step (Start server)
go run main.go

# Only below endpoints are supported as of now
- `/` - GET  
- `/echo/<message>` - GET  
- `/user-agent` - GET  
- `/files/<filename>` - GET & POST
