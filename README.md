# Knox
> AWS credential process helper



## About

Knox is a powerful utility designed to simplify and streamline the process of managing AWS credentials. Whether you're frequently switching between different AWS profiles or managing credentials issuance in an SSO environment, Knox provides a straightforward CLI tool to handle these tasks effortlessly. Commands like `knox select`, `knox creds select`, `knox creds last-used`, and `knox clean` make it easy to navigate and manipulate your AWS credential configurations. Its configurable nature, showcased in the `~/.aws/config` setup recommendations, ensures seamless integration into your AWS workflows. Whether you're in development, staging, or production, Knox helps maintain efficient and secure AWS credential management.

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
  curl -sL -o ./aws-knox_1.0.3_amd64.deb https://github.com/null93/aws-knox/releases/download/1.0.3/aws-knox_1.0.3_amd64.deb
  sudo dpkg -i ./aws-knox_1.0.3_amd64.deb
  rm ./aws-knox_1.0.3_amd64.deb
  ```

  ### arm64

  ```shell
  curl -sL -o ./aws-knox_1.0.3_arm64.deb https://github.com/null93/aws-knox/releases/download/1.0.3/aws-knox_1.0.3_arm64.deb
  sudo dpkg -i ./aws-knox_1.0.3_arm64.deb
  rm ./aws-knox_1.0.3_arm64.deb
  ```
</details>

<details>
  <summary>Red Hat</summary>

  ### aarch64

  ```shell
  rpm -i https://github.com/null93/aws-knox/releases/download/1.0.3/aws-knox-1.0.3-1.aarch64.rpm
  ```

  ### x86_64

  ```shell
  rpm -i https://github.com/null93/aws-knox/releases/download/1.0.3/aws-knox-1.0.3-1.x86_64.rpm
  ```
</details>

<details>
  <summary>Alpine</summary>

  ### aarch64

  ```shell
  curl -sL -o ./aws-knox_1.0.3_aarch64.apk https://github.com/null93/aws-knox/releases/download/1.0.3/aws-knox_1.0.3_aarch64.apk
  apk add --allow-untrusted ./aws-knox_1.0.3_aarch64.apk
  rm ./aws-knox_1.0.3_aarch64.apk
  ```

  ### x86_64

  ```shell
  curl -sL -o ./aws-knox_1.0.3_x86_64.apk https://github.com/null93/aws-knox/releases/download/1.0.3/aws-knox_1.0.3_x86_64.apk
  apk add --allow-untrusted ./aws-knox_1.0.3_x86_64.apk
  rm ./aws-knox_1.0.3_x86_64.apk
  ```
</details>

<details>
  <summary>Arch</summary>

  ### aarch64

  ```shell
  curl -sL -o ./aws-knox-1.0.3-1-aarch64.pkg.tar.zst https://github.com/null93/aws-knox/releases/download/1.0.3/aws-knox-1.0.3-1-aarch64.pkg.tar.zst
  sudo pacman -U ./aws-knox-1.0.3-1-aarch64.pkg.tar.zst
  rm ./aws-knox-1.0.3-1-aarch64.pkg.tar.zst
  ```

  ### x86_64

  ```shell
  curl -sL -o ./aws-knox-1.0.3-1-x86_64.pkg.tar.zst https://github.com/null93/aws-knox/releases/download/1.0.3/aws-knox-1.0.3-1-x86_64.pkg.tar.zst
  sudo pacman -U ./aws-knox-1.0.3-1-x86_64.pkg.tar.zst
  rm ./aws-knox-1.0.3-1-x86_64.pkg.tar.zst
  ```
</details>

## Setup

Recommended configuration for `~/.aws/config`, feel free to swap out the commands to suit your needs:

```ini
[default]
region = us-east-1
output = json
credential_process = knox creds select

[profile last]
region = us-east-1
output = json
credential_process = knox creds last-used

[profile pick]
region = us-east-1
output = json
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

Here is another use-case where this tool can come in handy. If you use SSM Session Manager to SSH into your EC2 instances, you can use knox to switch between different AWS profiles and start an interactive session with a specific instance. Here is an example of how you can achieve this:

```shell
function ssh-aws () {
    if [[ $# -ne 1 ]]; then
        echo "Usage: ssh-aws <instance-id>"
        return 1
    fi
    aws --profile pick ssm start-session --target $1 --document-name AWS-StartInteractiveCommand --parameters command="sudo su - \`id -un 9001\`"
}
```

Now you can SSH into an EC2 instance using the following command:

```
$ ssh-aws i-00000000000000000
```
