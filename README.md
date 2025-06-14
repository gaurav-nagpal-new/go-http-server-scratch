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
/ - GET \n
/echo/123 - GET \n
/user-agent - GET \n
/files/file_123 - GET & POST
