cd bin

echo starting IDIPServer
IDIPServer.exe
ping -n 1 127.1 >nul

echo start all done

cd ..

pause