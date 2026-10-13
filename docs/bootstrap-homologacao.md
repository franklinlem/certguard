# Administrador inicial da homologação

E-mail informado pelo responsável: `franklinlem@gmail.com`. A senha deve ser digitada pelo responsável no terminal seguro do Coolify; não enviar na conversa, não inserir no Git ou como argumento de linha de comando. CLI aceita senha com no mínimo 10 caracteres via stdin e permite somente o primeiro administrador.

A orientação para criar a senha é: **Escolha uma senha com no mínimo 10 caracteres. Você pode usar letras, números e símbolos.** Senhas muito longas podem exigir redução. A nova regra depende da atualização da aplicação; o administrador existente não precisa ser recriado.

Após banco/API saudáveis, abra o terminal do servidor `localhost` em [Coolify](https://coolify.franklem.uk/server/kuo8x8rohq7hfhibvaybw7cb/terminal). Confirme hostname `vmi3570746` e cole o bloco abaixo. Ele identifica exclusivamente a API do recurso CertGuard, exige correspondência única e pede a senha duas vezes sem exibi-la. Não cria senha padrão.

```bash
(
  set +x
  set -eu
  test "$(hostname)" = vmi3570746 || exit 1
  CG_CONTAINER=$(docker ps --filter label=com.docker.compose.project=rk48swz6mdqduoyyhfnqhylp --filter label=com.docker.compose.service=api --format '{{.ID}}')
  test -n "$CG_CONTAINER" || { echo 'API CertGuard não encontrada.'; exit 1; }
  test "$(printf '%s\n' "$CG_CONTAINER" | wc -l)" -eq 1 || exit 1
  IFS= read -r -s -p 'Senha do administrador (mínimo 10 caracteres): ' CG_PASSWORD
  printf '\n'
  IFS= read -r -s -p 'Repita a senha: ' CG_CONFIRM
  printf '\n'
  test "$CG_PASSWORD" = "$CG_CONFIRM" || { echo 'Senhas diferentes.'; exit 1; }
  printf '%s' "$CG_PASSWORD" | docker exec -i "$CG_CONTAINER" certguard admin-create franklinlem@gmail.com
  unset CG_PASSWORD CG_CONFIRM
)
```

Se houver administrador existente, a CLI recusa o bootstrap. Não apague usuários ou banco para contornar isso. Depois entre somente em `https://certguard-dev.franklem.uk`, guarde a senha no seu gerenciador e valide login/logout. Esse procedimento não configura backup, não habilita scheduler/entregas e não altera produção.
