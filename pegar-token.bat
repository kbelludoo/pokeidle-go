@echo off
chcp 65001 > nul
powershell -ExecutionPolicy Bypass -File "%~dp0scripts\pegar-token.ps1"
pause
