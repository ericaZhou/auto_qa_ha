DIR="$( cd "$( dirname "$0" )" && pwd )"

cd $DIR/src/auto-qa/nos-cli/agent/main 

# CGO_ENABLED=0 GOOS=windows GOARCH=386 go install

# mv $DIR/bin/windows_386/main.exe $DIR/dist/agent_win32.exe

# CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go install

# mv $DIR/bin/windows_amd64/main.exe $DIR/dist/agent_win64.exe

# CGO_ENABLED=0 GOOS=linux GOARCH=386 go install

# mv $DIR/bin/linux_386/main $DIR/dist/agent_linux32`

CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go install

mv $dirname/bin/linux_amd64/main $DIR/dist/agent_linux64

test