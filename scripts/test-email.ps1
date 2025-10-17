param(
  [string]$Email = "",
  [string]$Name = "Franco",
  [string]$Password = "demopass1",
  [string]$ApiBase = "http://localhost:8080/api/v1"
)

if (-not $Email) { Write-Error "Debes pasar -Email 'tu@correo.com'"; exit 1 }

$bodyObj = [pscustomobject]@{ name = $Name; email = $Email; password = $Password }
$bodyJson = $bodyObj | ConvertTo-Json -Compress

Write-Host "Registrando:" $Email
try {
  $resp = Invoke-RestMethod -Uri "$ApiBase/users/register" -Method Post -Body $bodyJson -ContentType 'application/json'
  Write-Host "REGISTER OK"
  $resp | ConvertTo-Json -Compress
} catch {
  $status = $_.Exception.Response.StatusCode.value__
  Write-Host "REGISTER FAIL - STATUS:" $status
  if ($_.Exception.Response) {
    $sr = New-Object System.IO.StreamReader ($_.Exception.Response.GetResponseStream())
    $sr.ReadToEnd()
  }
}
