FROM rust:latest as builder
WORKDIR /usr/src

# Create placeholder project
RUN USER=root cargo new --bin rust-chain-hyper
WORKDIR /usr/src/rust-chain-hyper
RUN mv src/main.rs src/rust-chain-hyper.rs
COPY Cargo.toml Cargo.lock ./
RUN cargo build --release
RUN rm src/*.rs

# Add my files and build the executable
ADD . ./
RUN ls -al target/release/deps
RUN rm ./target/release/deps/rust_chain_hyper*
RUN cargo build --release

# Move executable to run container
FROM debian:buster-slim
RUN apt-get update && rm -rf /var/lib/apt/lists/*
COPY --from=builder /usr/src/rust-chain-hyper/target/release/rust-chain-hyper ./rust-chain-hyper
CMD ["./rust-chain-hyper"]