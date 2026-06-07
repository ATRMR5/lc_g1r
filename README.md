1] Install Go https://go.dev/dl/<br>

2] Get **lockcracker.go** from repo<br>

3] Open CMD<br>

4] **Set the path to** the root directory of the folder where **lockcracker.go** is located<br>
		for example: the file is placed in C:\Users\%username%\Desktop\lockcracker.go<br>
		then set path in CMD to the C:\Users\%username%\Desktop<br>
		
5] Type into CMD:<br>
`go build -ldflags="-s -w" -o lockcracker.exe lockcracker.go`<br>

6] **lockcracker.exe** should appear on your desktop and ready to use<br>
