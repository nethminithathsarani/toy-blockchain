$projectPath = $PSScriptRoot

Start-Process powershell -ArgumentList @(
    "-NoExit",
    "-Command",
    "Set-Location '$projectPath'; go run . node -address localhost:8001 -peers localhost:8002,localhost:8003"
)

Start-Sleep -Seconds 2

Start-Process powershell -ArgumentList @(
    "-NoExit",
    "-Command",
    "Set-Location '$projectPath'; go run . node -address localhost:8002 -peers localhost:8001,localhost:8003"
)

Start-Sleep -Seconds 2

Start-Process powershell -ArgumentList @(
    "-NoExit",
    "-Command",
    "Set-Location '$projectPath'; go run . node -address localhost:8003 -peers localhost:8001,localhost:8002"
)