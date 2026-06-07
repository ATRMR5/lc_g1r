- Install Go https://go.dev/dl/
- Get **lockcracker.go** from repo
- Open CMD 
- **Set the path to** the root directory of the folder where **lockcracker.go** is located
  for example the file is placed in "C:\Users\%username%\Desktop\lockcracker.go"
  then set path in CMD to the "C:\Users\%username%\Desktop"
- Type into CMD: **go build -ldflags="-s -w" -o lockcracker.exe lockcracker.go**
- **lockcracker.exe** should appear on your desktop
