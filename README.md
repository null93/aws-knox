# Knox
> AWS credential process helper

```
$ knox select
$ knox creds select
$ knox creds last-used
$ knox clean creds sso -a
```

## Install

<details>
  <summary>Darwin</summary>

  ### Intel & ARM

  ```shell
  brew tap null93/tap
  brew install aws-knox
  ```
</details>

<details>
  <summary>Debian</summary>

  ### amd64

  ```shell
  curl -sL -o ./aws-knox_1.0.0_amd64.deb https://github.com/null93/aws-knox/releases/download/1.0.0/aws-knox_1.0.0_amd64.deb
  sudo dpkg -i ./aws-knox_1.0.0_amd64.deb
  rm ./aws-knox_1.0.0_amd64.deb
  ```

  ### arm64

  ```shell
  curl -sL -o ./aws-knox_1.0.0_arm64.deb https://github.com/null93/aws-knox/releases/download/1.0.0/aws-knox_1.0.0_arm64.deb
  sudo dpkg -i ./aws-knox_1.0.0_arm64.deb
  rm ./aws-knox_1.0.0_arm64.deb
  ```
</details>

<details>
  <summary>Red Hat</summary>

  ### aarch64

  ```shell
  rpm -i https://github.com/null93/aws-knox/releases/download/1.0.0/aws-knox-1.0.0-1.aarch64.rpm
  ```

  ### x86_64

  ```shell
  rpm -i https://github.com/null93/aws-knox/releases/download/1.0.0/aws-knox-1.0.0-1.x86_64.rpm
  ```
</details>

<details>
  <summary>Alpine</summary>

  ### aarch64

  ```shell
  curl -sL -o ./aws-knox_1.0.0_aarch64.apk https://github.com/null93/aws-knox/releases/download/1.0.0/aws-knox_1.0.0_aarch64.apk
  apk add --allow-untrusted ./aws-knox_1.0.0_aarch64.apk
  rm ./aws-knox_1.0.0_aarch64.apk
  ```

  ### x86_64

  ```shell
  curl -sL -o ./aws-knox_1.0.0_x86_64.apk https://github.com/null93/aws-knox/releases/download/1.0.0/aws-knox_1.0.0_x86_64.apk
  apk add --allow-untrusted ./aws-knox_1.0.0_x86_64.apk
  rm ./aws-knox_1.0.0_x86_64.apk
  ```
</details>

<details>
  <summary>Arch</summary>

  ### aarch64

  ```shell
  curl -sL -o ./aws-knox-1.0.0-1-aarch64.pkg.tar.zst https://github.com/null93/aws-knox/releases/download/1.0.0/aws-knox-1.0.0-1-aarch64.pkg.tar.zst
  sudo pacman -U ./aws-knox-1.0.0-1-aarch64.pkg.tar.zst
  rm ./aws-knox-1.0.0-1-aarch64.pkg.tar.zst
  ```

  ### x86_64

  ```shell
  curl -sL -o ./aws-knox-1.0.0-1-x86_64.pkg.tar.zst https://github.com/null93/aws-knox/releases/download/1.0.0/aws-knox-1.0.0-1-x86_64.pkg.tar.zst
  sudo pacman -U ./aws-knox-1.0.0-1-x86_64.pkg.tar.zst
  rm ./aws-knox-1.0.0-1-x86_64.pkg.tar.zst
  ```
</details>

## Setup

Recommended configuration for `~/.aws/config`:

```
[default]
credential_process = knox creds select

[profile last]
credential_process = knox creds last-used

[profile pick]
credential_process = knox select

[sso-session development-sso]
sso_region = us-east-1
sso_registration_scopes = sso:account:access
sso_start_url = https://d-2222222222.awsapps.com/start

[sso-session staging-sso]
sso_region = us-east-1
sso_registration_scopes = sso:account:access
sso_start_url = https://d-1111111111.awsapps.com/start

[sso-session production-sso]
sso_region = us-east-1
sso_registration_scopes = sso:account:access
sso_start_url = https://d-0000000000.awsapps.com/start
```

## Example

```
function ssh-aws () {
    if [[ $# -ne 1 ]]; then
        echo "Usage: ssh-aws <instance-id>"
        return 1
    fi
    aws --profile pick ssm start-session --target $1 --document-name AWS-StartInteractiveCommand --parameters command="sudo su - \`id -un 9001\`"
}
```

```
ssh-aws i-00000000000000000
```
