# ADR-003 — Inspeção de certificados inválidos
Status: aceito. Data: 12/10/2026.

O monitor precisa observar certificados expirados ou não confiáveis. Um fluxo exclusivo de inspeção usa InsecureSkipVerify=true, imediatamente seguido por VerifyHostname e x509.Verify com raízes do sistema, intermediárias recebidas, tempo atual e uso ServerAuth. Nenhum dado de aplicação, cookie ou credencial é transmitido. O cliente TLS não é exposto nem reutilizado.

Handshake concluído nunca define confiança. HostnameValid e ChainValid são independentes; cadeia sem hostname ainda verifica datas, assinaturas, uso e restrições. Resultado prioriza validade futura, expiração, hostname e cadeia; os booleanos preservam falhas simultâneas. Sem OCSP/CRL e sem download AIA: não alegar detecção de revogação.

TLS mínimo 1.2; servidores exclusivamente legados resultam em tls_error. Timeout total de 10 segundos por padrão, máximo 30 na construção; deadline TCP/TLS e contexto DNS. Dados do handshake são limitados pela implementação crypto/tls do Go. Limites de concorrência e quotas serão introduzidos com o scheduler antes de expor checks HTTP.

Dias restantes usam floor, logo um certificado recém-expirado não aparece com zero positivo. Severidade usa duração exata: <=7 dias critical, <=30 warning; certificado expirado é critical. Em falhas anteriores ao certificado os booleanos e dias são null/ausentes.

Referências: https://pkg.go.dev/crypto/tls e https://pkg.go.dev/crypto/x509.
