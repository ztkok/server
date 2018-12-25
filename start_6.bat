cd bin

echo starting DataCenter
DataCenter.exe
ping -n 1 127.1 >nul

echo start all done

cd ..

pause