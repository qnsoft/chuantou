@echo off
set target=cn_ips.txt
set source=delegated-apnic-latest.txt

echo checking...
if not exist %source% (
  echo could not found source file %source%
  echo please download delegated-apnic-latest.txt from http://ftp.apnic.net/apnic/stats/apnic/delegated-apnic-latest
) else (
  echo processing...
  if exist %target% ( del /F /Q %target% )
  for /F "tokens=4-5 delims=|" %%a in ('findstr /c:"apnic|CN|ipv4|" %source%') do echo %%a,%%b>> %target%
  echo finished
)