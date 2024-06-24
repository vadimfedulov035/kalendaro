#!/bin/bash

# directory to store user info
CONF_DIR="/etc/ifc-website"
mkdir "$CONF_DIR"
cp -r "/root/ifc-website" "/etc"

# files to store user info
WEBHOST_FILE="$CONF_DIR/webhost.txt"
EMAIL_FILE="$CONF_DIR/email.txt"

####################################################################################################
#                      __     ______  ____    ____  _____ _____ _   _ ____                         #
#                      \ \   / /  _ \/ ___|  / ___|| ____|_   _| | | |  _ \                        #
#                       \ \ / /| | | \___ \  \___ \|  _|   | | | | | | |_) |                       #
#                        \ V / | |_| |___) |  ___) | |___  | | | |_| |  __/                        #
#                         \_/  |____/|____/  |____/|_____| |_|  \___/|_|                           #
####################################################################################################

if [ ! -d "/root/vds-setup" ]; then
        mkdir "/root/vds-setup"
        git clone "git@github.com:vadimfedulov035/vds-setup.git" "/root/vds-setup"
fi

source "/root/vds-setup/vars.sh"
source "/root/vds-setup/sys.sh"

get_vars "webhost email"

cron_cmds=(
	"*/5 * * * * /usr/local/bin/datilo"
	"0 * * * * [ \$(df --output=pcent /dev/sda1 | tr -dc '0-9') -gt 95 ] && /usr/local/bin/purigilo"
	"0 0 1 * * certbot renew"
)

set_vds "${cron_cmds[@]}"

####################################################################################################
#                 ____  _____ ____  ____    ___ _   _ ____ _____  _    _     _                     #
#                |  _ \| ____|  _ \/ ___|  |_ _| \ | / ___|_   _|/ \  | |   | |                    #
#                | | | |  _| | |_) \___ \   | ||  \| \___ \ | | / _ \ | |   | |                    #
#                | |_| | |___|  __/ ___) |  | || |\  |___) || |/ ___ \| |___| |___                 #
#                |____/|_____|_|   |____/  |___|_| \_|____/ |_/_/   \_\_____|_____|                #
####################################################################################################

figlet "DEPS INSTALL"

install_deps() {
	export DEBIAN_FRONTEND=noninteractive
	apt update && apt upgrade -y
	apt install locales -y
	apt install nginx certbot python3-certbot-nginx -y
	apt install texlive texlive-xetex -y
	apt install poppler-utils cron figlet -y
	apt autoremove -y && apt clean -y
}

install_deps

####################################################################################################
#               ____  _____ ______     _______ ____    ____  _____ _____ _   _ ____                #
#              / ___|| ____|  _ \ \   / / ____|  _ \  / ___|| ____|_   _| | | |  _ \               #
#              \___ \|  _| | |_) \ \ / /|  _| | |_) | \___ \|  _|   | | | | | | |_) |              #
#               ___) | |___|  _ < \ V / | |___|  _ <   ___) | |___  | | | |_| |  __/               #
#              |____/|_____|_| \_\ \_/  |_____|_| \_\ |____/|_____| |_|  \___/|_|                  #
####################################################################################################

figlet "SERVER SETUP"

set_website() {
	# set bin
	cp /root/ifc-website/bin/datilo /root/ifc-website/bin/purigilo /usr/local/bin
	# set config
	cp conf/website.conf.default website.conf
	sed -i "s/myserver\.tld/$webhost/g" website.conf
	mv website.conf /etc/nginx/sites-available/kalendaro
	ln -s /etc/nginx/sites-available/kalendaro /etc/nginx/sites-enabled/
	# delete default
	rm -f /etc/nginx/sites-enabled/default
	# get certificates
	certbot -d "$webhost" --nginx -n --agree-tos --email "$email"
	systemctl enable nginx
	systemctl restart nginx
	mkdir /var/www/kalendaro
}

set_website
