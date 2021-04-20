# GO Upload Server - Command Line Tool
[![Go Reference](https://pkg.go.dev/badge/github.com/guilhermerodrigues680/gouploadserver.svg)](https://pkg.go.dev/github.com/guilhermerodrigues680/gouploadserver)

O **GO Upload Server** foi escrito para ser agil e permitir a inicialização rápida de um servidor de arquivos a partir de um diretório.

O **GO Upload Server** faz leitura e escrita usando buffers o que faz ele consumir pouquissima memoria (small memory footprint).

*Nota: Embora **GO Upload Server** seja uma ótima maneira de servir facilmente arquivos de um diretório, ele não deve ser usado em um ambiente de produção pois ele não implementa verificações de segurança.*

## Motivações para o projeto
Servidores da web amplamente usados, como NGINX, Apache e Tomcat, são excelentes porém consomem muito tempo para serem configurados.

O GO possui o package `http` com a `func FileServer` que inicia um servidor de arquivos a partir de um diretório, porém o `FileServer` não permite customizações. Ex: Alteração no template HTML padrão e CSS personalizado.

Então para ter um servidor personalizável, com novas funcionalidades e agil este projeto foi desenvolvido como uma Command Line Tool.

## Features
- Servidor de arquivos.
- Servidor websites pois implementa  MIME types.
- Baixíssimo consumo de memória.
- Alteração fácil da porta do servidor via flag
- Navegador de arquivos com opção para upload de arquivo no diretório navegado.
- Implementa o renomeio dos arquivos enviados para não sobreescrever os arquivos originais do diretório (pode ser desativado via flag).
- Usa o Go templates internamente permitindo a customização do navegador de arquivos.

## Instalação
### Instalação com Go (requer v1.16+).

```console
go get -u guilhermerodrigues680/gouploadserver
```

### Instalação com binários pré-compilados (qualquer sistema operacional) 
Para instalar a versão mais recente do gouploadserver a partir de binários pré-compilados, siga estas instruções:

1. Baixe manualmente em [github.com/guilhermerodrigues680/gouploadserver/releases](https://github.com/guilhermerodrigues680/gouploadserver/releases) o arquivo zip correspondente ao seu sistema operacional e arquitetura do computador (gouploadserver-<version>-<os><arch>.zip), ou baixe o arquivo usando comandos como os seguintes:

```sh
$ PR_REL="https://github.com/guilhermerodrigues680/gouploadserver/releases"
$ curl -LO $PR_REL/download/v1.0.0/gouploadserver-v1.0.0-linux-amd64.zip
```

2. Descompacte o arquivo em `$HOME/.local` ou em um diretório de sua escolha. Por exemplo:

```sh
$ unzip gouploadserver-v1.0.0-linux-amd64.zip -d $HOME/.local
```

3. Atualize os seu `PATH` para incluir o caminho para o executável gouploadserver. Por exemplo: 

```sh
$ export PATH="$PATH:$HOME/.local/bin"
```

## Como usar

Start do **gouploadserver** com as configurações padrão:

```console
$ gouploadserver
```

É possivel passar flags para o **gouploadserver**:

```sh
# Usage: gouploadserver [options] [path]
$ gouploadserver --port 8082 ./folder
```

### Command-Line Options
```console
Usage: gouploadserver [options] [path]
[path] defaults to ./
Options are:
  --dev                      Use development settings (default false)
  --keep-upload-filename     Keep original upload file name: Use 'filename.ext' instead of 'filename<-random>.ext' (default false)
  --port                     Port to use (default 8000)
  --spa                      Return to all files not found /index.html (default false)
  --version                  Show version number and quit (default false)
  --watch-mem                Watch memory usage (default false)
  --help                     Display usage information (this message)
  -h                         Display usage information (this message) (shorthand)
```

## Configuração do projeto para desenvolvimento

Requer o GO v1.16+

Use o make para 

```sh
make build

# run:
./bin/gouploadserver
```

### Installation

```sh
make install

# run:
gouploadserver
```

<details>
<summary><h2>References</h2></summary>
<br>
- https://www.digitalocean.com/community/tutorials/how-to-build-and-install-go-programs-pt
- https://golang.org/doc/tutorial/compile-install
- https://golang.org/ref/mod#go-install
- https://makefiletutorial.com/

```sh
➜  cmd go list -f '{{.Target}}'
/Users/guilherme/go/bin/cmd
```

```sh
go test -v -benchmem -bench=.
```
date -u +"%Y%m%d%H%M%S"
TZ=UTC date +"%Y%m%d%H%M%S"
TZ=GMT date +"%Y%m%d%H%M%S"

https://pkg.go.dev/github.com/guilhermerodrigues680/gouploadserver

git tag v0.0.0-alpha.0-$(date -u +"%Y%m%d%H%M%S")

</details>
