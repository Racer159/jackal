# jackal-injector

A tiny (<1MiB) binary statically-linked with musl in order to fit as a configmap

## Building on Ubuntu

```bash
# install rust
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y --no-modify-path
source $HOME/.cargo/env

# install build-essential
sudo apt install build-essential -y

# build w/ musl
rustup target add x86_64-unknown-linux-musl
cargo build --target x86_64-unknown-linux-musl --release
```

## Checking Binary Size

Due to the ConfigMap size limit (1MiB for binary data), we need to make sure the binary is small enough to fit.

```bash
cargo build --target x86_64-unknown-linux-musl --release

cargo build --target aarch64-unknown-linux-musl --release

size_linux=$(du --si target/x86_64-unknown-linux-musl/release/jackal-injector | cut -f1)
echo "Linux binary size: $size_linux"
size_aarch64=$(du --si target/aarch64-unknown-linux-musl/release/jackal-injector | cut -f1)
echo "aarch64 binary size: $size_aarch64"
```
