# Base da interface

Página informativa em pt-BR, responsiva e acessível, sem dados fictícios. Não é um dashboard funcional. Fontes locais do sistema, sem imagens ou recursos externos e sem mudanças de layout causadas por mídia sem dimensões. Respeita preferência de tema claro/escuro.

Homologação e dashboard ficam fora de indexação (meta robots e X-Robots-Tag); robots.txt também restringe rastreamento, mas não substitui autenticação. Uma landing pública com canonical, sitemap e dados estruturados ficará em incremento próprio.

Assets com hash recebem cache de um ano. HTML revalida para evitar referências antigas. Headers de segurança no Nginx. HSTS pertence ao proxy após validação HTTPS.

Referências consultadas:
- https://developers.google.com/search/docs?hl=pt-br
- https://developer.chrome.com/docs/performance/insights/cache?hl=pt-br
- https://web.dev/articles/optimize-cls?hl=pt-br#images_without_dimensions
