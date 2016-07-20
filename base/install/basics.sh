#!/bin/bash

echo "installing basics"
apt-get update
apt-get install -y \
  sudo  \
  build-essential \
  curl \
  gcc \
  make \
  openssl \
  software-properties-common \
  wget \
  nano \
  unzip \
  libxslt-dev \
  libxml2-dev
