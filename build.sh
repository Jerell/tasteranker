#!/bin/bash

# Ensure the assets/js directory exists
mkdir -p assets/js

# Bundle and compile TypeScript with Bun
bun build ts/index.ts --outdir assets/js 

