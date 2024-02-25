# FROM balenalib/raspberrypi4-64-debian
FROM --platform=linux/arm64  balenalib/aarch64-debian:bookworm
ARG ARCITECTURE="arm64"
ARG SERVICE_NAME="hmi"

RUN mkdir -p /mnt/data
RUN apt update

# psmisc: For kill all
# tmux: To be able to reconnect to the session

# Devtools
RUN apt-get install -y vim \
        wget \
        nfs-common \
        psmisc \ 
        tmux \
        iputils-ping \
        net-tools \
        avahi-daemon \
        cifs-utils \
        sshfs

# packages for compulab devices
RUN apt-get install -y acpi-support-base acpid gpiod 

RUN wget https://raw.githubusercontent.com/gdraheim/docker-systemctl-replacement/master/files/docker/systemctl3.py -O /usr/local/bin/systemctl
RUN chmod u+x /usr/local/bin/systemctl


# COPY ./service/docker/startup /usr/local/startup
# RUN chmod +x /usr/local/startup/startup.sh
# ENTRYPOINT [ "/usr/local/startup/startup.sh" ]
CMD ["sh"]