@echo off
title KafePlat O'rnatuvchi
color 0A
echo ==============================================
echo        KafePlat POS Tizimini O'rnatish
echo ==============================================
echo.
echo Bu dastur kompyuteringiz (Ish stoli) ga KafePlat kassa 
echo dasturini o'rnatadi.
echo.
pause

set SCRIPT="%TEMP%\%RANDOM%-%RANDOM%-%RANDOM%-%RANDOM%.vbs"

echo Set oWS = WScript.CreateObject("WScript.Shell") >> %SCRIPT%
echo sLinkFile = "%USERPROFILE%\Desktop\KafePlat.lnk" >> %SCRIPT%
echo Set oLink = oWS.CreateShortcut(sLinkFile) >> %SCRIPT%
echo oLink.TargetPath = "chrome.exe" >> %SCRIPT%
echo oLink.Arguments = "--app=https://kafe.securehub.uz" >> %SCRIPT%
echo oLink.Description = "KafePlat POS System" >> %SCRIPT%
echo oLink.IconLocation = "chrome.exe, 0" >> %SCRIPT%
echo oLink.Save >> %SCRIPT%

cscript /nologo %SCRIPT%
del %SCRIPT%

echo.
echo ==============================================
echo MUVAFFAQIYATLI O'RNATILDI!
echo ==============================================
echo Ish stolingizda (Рабочий стол) "KafePlat" nomli 
echo ikonka paydo bo'ldi. Uni bosib dasturga kirishingiz mumkin.
echo.
pause
