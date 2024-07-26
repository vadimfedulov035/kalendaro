#!/bin/bash

###############################################################################
#            __     ______  ____    ____  _____ _____ _   _ ____              #
#            \ \   / /  _ \/ ___|  / ___|| ____|_   _| | | |  _ \             #
#             \ \ / /| | | \___ \  \___ \|  _|   | | | | | | |_) |            #
#              \ V / | |_| |___) |  ___) | |___  | | | |_| |  __/             #
#               \_/  |____/|____/  |____/|_____| |_|  \___/|_|                #
###############################################################################

CONF_DIR="/etc/${PWD##*/}"
mkdir -p $CONF_DIR

set_conf() {
        var=$1
        filename=$2
        file="$CONF_DIR/$filename"
        while [ ! -s "$file" ]; do
                read -p "Type $var: " answer
                if [[ -n "$answer" ]]; then
                        echo "$answer" > "$file"
                        echo "${var^} saved to file '$file'"

                else
                        echo "No $var provided!"
                fi
        done
}

set_conf "webhost" "webhost.txt"
set_conf "email" "email.txt"

webhost=$(cat "$CONF_DIR/webhost.txt")
email=$(cat "$CONF_DIR/email.txt")

###############################################################################
#      ____  _____ ____  ____    ___ _   _ ____ _____  _    _     _           #
#     |  _ \| ____|  _ \/ ___|  |_ _| \ | / ___|_   _|/ \  | |   | |          #
#     | | | |  _| | |_) \___ \   | ||  \| \___ \ | | / _ \ | |   | |          #
#     | |_| | |___|  __/ ___) |  | || |\  |___) || |/ ___ \| |___| |___       #
#     |____/|_____|_|   |____/  |___|_| \_|____/ |_/_/   \_\_____|_____|      #
###############################################################################

export DEBIAN_FRONTEND=noninteractive
apt update && apt upgrade -y
apt install locales -y
apt install nginx certbot python3-certbot-nginx -y
apt install texlive texlive-xetex -y
apt install cron poppler-utils figlet -y
apt autoremove -y && apt clean -y

###############################################################################
#    ____  _____ ______     _______ ____    ____  _____ _____ _   _ ____      #
#   / ___|| ____|  _ \ \   / / ____|  _ \  / ___|| ____|_   _| | | |  _ \     #
#   \___ \|  _| | |_) \ \ / /|  _| | |_) | \___ \|  _|   | | | | | | |_) |    #
#    ___) | |___|  _ < \ V / | |___|  _ <   ___) | |___  | | | |_| |  __/     #
#   |____/|_____|_| \_\ \_/  |_____|_| \_\ |____/|_____| |_|  \___/|_|        #
###############################################################################

# set bin
cp /root/kalendaro/bin/datilo /root/kalendaro/bin/purigilo /usr/local/bin
# set conf
cp conf/website.conf.default website.conf
sed -i "s/myserver\.tld/$webhost/g" website.conf
mv website.conf /etc/nginx/sites-available/kalendaro
ln -s /etc/nginx/sites-available/kalendaro /etc/nginx/sites-enabled/
rm -f /etc/nginx/sites-enabled/default
# get certs
certbot -d "$webhost" --nginx -n --agree-tos --email "$email"
# restart nginx
systemctl enable nginx
systemctl restart nginx
# prepare webhost
mkdir /var/www/kalendaro

###############################################################################
#            ____ ____   ___  _   _   ____  _____ _____ _   _ ____            #
#           / ___|  _ \ / _ \| \ | | / ___|| ____|_   _| | | |  _ \           #
#          | |   | |_) | | | |  \| | \___ \|  _|   | | | | | | |_) |          #
#          | |___|  _ <| |_| | |\  |  ___) | |___  | | | |_| |  __/           #
#           \____|_| \_\\___/|_| \_| |____/|_____| |_|  \___/|_|              #
###############################################################################

cp /root/kalendaro/conf/crontab /etc/cron.d/kalendaro
systemctl restart cron
