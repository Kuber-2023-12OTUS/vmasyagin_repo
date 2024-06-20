CONTAINERD_VERSION=1.7.18
RUNC_VERSION=v1.1.12
CNI_PLUGINS_VERSION=v1.5.0
OS=linux
ARCH=amd64
K8S_VERSION_MINOR=1.29

cat <<EOF | sudo tee /etc/modules-load.d/k8s.conf
overlay
br_netfilter
EOF

# Enable IPv4 packet forwarding
sudo modprobe overlay
sudo modprobe br_netfilter

# sysctl params required by setup, params persist across reboots
cat <<EOF | sudo tee /etc/sysctl.d/k8s.conf
net.bridge.bridge-nf-call-iptables  = 1
net.bridge.bridge-nf-call-ip6tables = 1
net.ipv4.ip_forward                 = 1
EOF

# Apply sysctl params without reboot
sudo sysctl --system

# Disable swap
sudo swapoff -a
(crontab -l 2>/dev/null; echo "@reboot /sbin/swapoff -a") | crontab - || true

# Install containerd
curl -s -L -O https://github.com/containerd/containerd/releases/download/v$CONTAINERD_VERSION/containerd-$CONTAINERD_VERSION-$OS-$ARCH.tar.gz
curl -s -L -O https://github.com/containerd/containerd/releases/download/v$CONTAINERD_VERSION/containerd-$CONTAINERD_VERSION-$OS-$ARCH.tar.gz.sha256sum
echo "$(cat containerd-$CONTAINERD_VERSION-$OS-$ARCH.tar.gz.sha256sum)" | sha256sum --check
tar Cxzvf /usr/local containerd-$CONTAINERD_VERSION-$OS-$ARCH.tar.gz
curl -s -L -o /lib/systemd/system/containerd.service https://raw.githubusercontent.com/containerd/containerd/main/containerd.service
systemctl daemon-reload
systemctl enable --now containerd

# Configuring the systemd cgroup driver
mkdir /etc/containerd
touch /etc/containerd/config.toml
containerd config default > /etc/containerd/config.toml
sed -i "s/SystemdCgroup = false/SystemdCgroup = true/" /etc/containerd/config.toml
systemctl restart containerd

# Install runc
curl -s -L -O https://github.com/opencontainers/runc/releases/download/$RUNC_VERSION/runc.$ARCH
curl -s -L -O https://github.com/opencontainers/runc/releases/download/$RUNC_VERSION/runc.sha256sum
echo "$(cat runc.sha256sum | grep $ARCH)" | sha256sum --check
install -m 755 runc.amd64 /usr/local/sbin/runc

# Install CNI plugins
curl -s -L -O https://github.com/containernetworking/plugins/releases/download/$CNI_PLUGINS_VERSION/cni-plugins-$OS-$ARCH-$CNI_PLUGINS_VERSION.tgz
curl -s -L -O https://github.com/containernetworking/plugins/releases/download/$CNI_PLUGINS_VERSION/cni-plugins-$OS-$ARCH-$CNI_PLUGINS_VERSION.tgz.sha256
echo "$(cat cni-plugins-$OS-$ARCH-$CNI_PLUGINS_VERSION.tgz.sha256)" | sha256sum --check
mkdir -p /opt/cni/bin
tar Cxzvf /opt/cni/bin cni-plugins-linux-amd64-v1.1.1.tgz

# Install kubeadm, kubelet and kubectl
sudo apt-get update
# apt-transport-https may be a dummy package; if so, you can skip that package
sudo apt-get install -y apt-transport-https ca-certificates curl gpg
# If the directory `/etc/apt/keyrings` does not exist, it should be created before the curl command, read the note below.
# sudo mkdir -p -m 755 /etc/apt/keyrings
curl -fsSL https://pkgs.k8s.io/core:/stable:/v${K8S_VERSION_MINOR}/deb/Release.key | sudo gpg --dearmor -o /etc/apt/keyrings/kubernetes-apt-keyring.gpg
# This overwrites any existing configuration in /etc/apt/sources.list.d/kubernetes.list
echo "deb [signed-by=/etc/apt/keyrings/kubernetes-apt-keyring.gpg] https://pkgs.k8s.io/core:/stable:/v${K8S_VERSION_MINOR}/deb/ /" | sudo tee /etc/apt/sources.list.d/kubernetes.list
sudo apt-get update
sudo apt-get install -y kubelet=1.29.0-1.1 kubeadm=1.29.0-1.1 kubectl=1.29.0-1.1
sudo apt-mark hold kubelet kubeadm kubectl
sudo systemctl enable --now kubelet