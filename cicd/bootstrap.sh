if [ ! -f ~/.ssh/id_rsa ]; then
  ssh-keygen -q -t rsa -N '' -f ~/.ssh/id_rsa
fi
if [ ! -f /opt/configs/elzion/release/gjcc_config.yml ]; then
  mkdir -p /opt/configs/elzion/release
  cp /app/configs/gjcc_config.yml /opt/configs/elzion/release/_config.yml
fi
/app/cicd
