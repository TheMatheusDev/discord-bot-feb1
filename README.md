# Discord Bot (feb-notify)

Este projeto é um bot para Discord escrito em Go, projetado com uma arquitetura limpa e focado em realizar um monitoramento periódico de um servidor de Squad usando a API pública `mysquadstats.com`.

## Funcionalidades

- **Slash Command `/next`**: Registra o usuário atual para um monitoramento que verifica a cada 5 minutos se o mapa atual do servidor de Squad foi trocado. Quando a troca é detectada, o usuário é mencionado com uma notificação do mapa anterior e o novo. O monitoramento encerra automaticamente após a notificação.

## Estrutura do Projeto

- `cmd/bot/`: Ponto de entrada (bootstrap) da aplicação. Contém o `main.go`.
- `internal/config/`: Gerencia o carregamento de variáveis de ambiente.
- `internal/squadstats/`: Cliente HTTP e estruturas de dados para interação com a API externa.
- `internal/monitor/`: Contém a regra de negócio do _polling_ a cada 5 minutos e o controle de monitoramentos ativos por usuário.
- `internal/bot/`: Contém a configuração, ciclo de vida e comandos da sessão do bot com o pacote `discordgo`.

## Como compilar e executar

1. Copie o arquivo de exemplo do ambiente e insira seu token:
   ```bash
   cp .env.example .env
   # Edite o .env para adicionar o DISCORD_TOKEN e (opcionalmente) o GUILD_ID
   ```

2. Baixe as dependências:
   ```bash
   go mod tidy
   ```

3. Compile e execute o bot:
   ```bash
   go run cmd/bot/main.go
   ```

4. No Discord, digite `/next` para iniciar o monitoramento.
