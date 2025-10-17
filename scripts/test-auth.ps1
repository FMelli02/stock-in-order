param(
  [string]$ApiBase = "http://localhost:8080/api/v1"
)

Write-Host "API Base:" $ApiBase

# Build unique email
$random = Get-Random
$email = "demo$random@example.com"
$name = "Usuario Demo"
$password = "demopass1"

# Prepare bodies
$regObj = [pscustomobject]@{ name = $name; email = $email; password = $password }
$regJson = $regObj | ConvertTo-Json -Compress

Write-Host "Registering:" $email
try {
  $registerResp = Invoke-RestMethod -Uri "$ApiBase/users/register" -Method Post -Body $regJson -ContentType 'application/json'
  Write-Host "REGISTER OK"
  $registerResp | ConvertTo-Json -Compress
} catch {
  Write-Host "REGISTER FAIL"
  Write-Host $_.Exception.Message
  if ($_.Exception.Response) {
    $sr = New-Object System.IO.StreamReader ($_.Exception.Response.GetResponseStream())
    $sr.ReadToEnd()
  }
}

# Login
$loginObj = [pscustomobject]@{ email = $email; password = $password }
$loginJson = $loginObj | ConvertTo-Json -Compress

try {
  $loginResp = Invoke-RestMethod -Uri "$ApiBase/users/login" -Method Post -Body $loginJson -ContentType 'application/json'
  Write-Host "LOGIN OK"
  $loginResp | ConvertTo-Json -Compress
} catch {
  Write-Host "LOGIN FAIL"
  Write-Host $_.Exception.Message
  if ($_.Exception.Response) {
    $sr2 = New-Object System.IO.StreamReader ($_.Exception.Response.GetResponseStream())
    $sr2.ReadToEnd()
  }
}