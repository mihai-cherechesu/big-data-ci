FROM ubuntu:20.04

RUN DEBIAN_FRONTEND=noninteractive apt update -y && DEBIAN_FRONTEND=noninteractive apt install -y apt-transport-https wget gcc g++ build-essential python3 python3-pip golang-go openjdk-17-jdk-headless

RUN wget https://packages.microsoft.com/config/ubuntu/20.04/packages-microsoft-prod.deb -O packages-microsoft-prod.deb

RUN dpkg -i packages-microsoft-prod.deb && rm packages-microsoft-prod.deb && apt update -y && apt install -y dotnet-sdk-6.0 aspnetcore-runtime-6.0 dotnet-runtime-6.0
