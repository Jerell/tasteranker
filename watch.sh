#!/bin/bash

# Ensure the assets/js directory exists
mkdir -p assets/js

# Start Bun in watch mode to bundle and compile TypeScript
bun build ts/index.ts --outdir assets/js --watch
