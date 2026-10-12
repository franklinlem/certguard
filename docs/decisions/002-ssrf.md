# ADR-002 — Saída TLS conservadora
Status: aceito. Data: 12/10/2026.

Aceitar apenas hostname DNS ASCII completo (IDN como punycode), porta 443, até 253 caracteres e 64 endereços DNS. IP literal, URL, credenciais, caminho, whitespace, nomes locais e sintaxe ambígua são rejeitados. Consulta com ponto final evita search domains do host.

Todos os A/AAAA devem passar pela política, inclusive os não escolhidos. Bloquear intervalos especiais IANA mesmo quando alguns têm exceções globais; permitir IPv6 global somente em 2000::/3 e bloquear suas reservas especiais. IPv4-mapped, NAT64 e 6to4 são bloqueados. Inclui CGNAT/Tailscale, metadata e endpoint de plataforma Azure.

Um IP por execução, revalidado e passado literalmente ao DialContext. Não resolver novamente no dial. Cada novo check resolve novamente. Sem retries nesta fase: registrar falha do primeiro IP, sem alegar saúde de todos os endereços. Sem proxies ou HTTP.

Testes podem substituir resolver/dial somente por campos privados no próprio pacote; nenhuma variável de ambiente permite desabilitar SSRF. Revisar registros IANA em atualizações da política.

Limite: a aplicação não controla roteamento/NAT. Um IP público roteado para um serviço local exige filtragem de egress e isolamento no host antes de expor o monitoramento. Regras de firewall dependem de inspeção e autorização da infraestrutura.

Referências consultadas:
- https://www.iana.org/assignments/iana-ipv4-special-registry/
- https://www.iana.org/assignments/iana-ipv6-special-registry/
