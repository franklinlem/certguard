# Interface — etapa 03

Interface administrativa em pt-BR: login, cadastro/edição, pausa/ativação, check manual, histórico paginado e arquivamento. O painel informa configuração do agendador e próxima execução prevista. Não há alertas, visão agregada ou listagem de arquivados; a UI informa esses limites. Histórico arquivado permanece consultável por id na API.

Tokens de sessão ficam em cookie HttpOnly; CSRF somente em memória. Erros acessíveis via role=alert, mensagens de conclusão via role=status, formulários com labels e estados por texto e cor. Fontes do sistema, tema conforme preferência, layout adaptável e sem mídia que cause deslocamento de layout.

Meta robots e X-Robots-Tag impedem indexação. Robots.txt também restringe rastreamento, mas não substitui autenticação. Landing pública, canonical, sitemap e dados estruturados serão um incremento próprio.

Assets com hash recebem cache de um ano, HTML revalida. CSP restringe recursos à própria origem. Em desenvolvimento Vite encaminha /api ao backend local; PUBLIC_APP_URL deve ser a origem exata da interface.

Referências consultadas:
- https://developers.google.com/search/docs?hl=pt-br
- https://developer.chrome.com/docs/performance/insights/cache?hl=pt-br
- https://web.dev/articles/optimize-cls?hl=pt-br#images_without_dimensions
