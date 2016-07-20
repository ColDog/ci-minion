#!/bin/bash -e

sudo gpg --keyserver hkp://keys.gnupg.net --recv-keys 409B6B1796C275462A1703113804BB82D39DC0E3

echo "installing rvm"
\curl -L https://get.rvm.io | bash -s stable

source /usr/local/rvm/scripts/rvm
rvm requirements

echo "installing ruby"
rvm install ruby
rvm use ruby --default
rvm rubygems current
gem install bundler
