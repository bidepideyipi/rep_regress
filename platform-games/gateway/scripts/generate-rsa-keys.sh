#!/bin/bash

set -e

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
KEYS_DIR="${SCRIPT_DIR}/../keys"

mkdir -p "$KEYS_DIR"

echo "Generating RSA-2048 key pair..."

openssl genrsa -out "$KEYS_DIR/private.pem" 2048
openssl rsa -in "$KEYS_DIR/private.pem" -pubout -out "$KEYS_DIR/public.pem"

chmod 600 "$KEYS_DIR/private.pem"
chmod 644 "$KEYS_DIR/public.pem"

echo "RSA keys generated:"
echo "  Private key: $KEYS_DIR/private.pem"
echo "  Public key:  $KEYS_DIR/public.pem"
