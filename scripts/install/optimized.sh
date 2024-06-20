#!/bin/bash
##################Gnu/Linux Kernel /tmp filesystem optimized####################h
set -e
uNames=`uname -s`
osName=${uNames: 0: 4}
if [ "$osName" == "Linu" ]
then
read -r -p "Are You Begin? [Y/n] " input
else
	echo "OS Error..."
	exit 1
fi
case "$input" in
    [yY][eE][sS]|[yY])
		echo "Yes"
                FSTAB=/etc/fstab
                if [ -f "${FSTAB}" ]; then
                  sed -i '$a\tmpfs        \/tmp        tmpfs        defaults     0 0' ${FSTAB}
                else
                    echo "The '${FSTAB}' file is not find"
                    exit 1
                fi
                SYSCTL=/etc/sysctl.conf
                if [ -f "${SYSCTL}" ]; then
                    main=`uname -r | awk -F . '{print $1}'` 
                    minor=`uname -r | awk -F . '{print $2}'` 
                        if [ "$main" -ge 4 ] && [ "$minor" -ge 9 ];then
                             echo "The Kernel bigger 4.9"
                             sed -i '$a\net.core.default_qdisc = fq' ${SYSCTL}
                             sed -i '$a\net.ipv4.tcp_congestion_control = bbr' ${SYSCTL}
                        else
                            echo "The Kernel less 4.9"
                        fi
                        sed -i '$a\net.ipv4.tcp_timestamps = 0' ${SYSCTL}
                        sed -i '$a\net.ipv4.tcp_synack_retries = 1' ${SYSCTL}
                        sed -i '$a\net.ipv4.tcp_fin_timeout = 15' ${SYSCTL}
                        sed -i '$a\net.ipv4.tcp_syn_retries = 2' ${SYSCTL}
                        sed -i '$a\net.ipv4.tcp_keepalive_time = 600' ${SYSCTL}
                        sed -i '$a\net.ipv4.tcp_orphan_retries = 3' ${SYSCTL}
                        sed -i '$a\net.ipv4.tcp_syncookies = 1' ${SYSCTL}
                        sed -i '$a\net.ipv4.tcp_tw_reuse = 0' ${SYSCTL}
                        sed -i '$a\net.ipv4.ip_local_port_range = 10240 65000' ${SYSCTL}
                        sed -i '$a\net.ipv4.tcp_max_syn_backlog = 262144' ${SYSCTL}
                        sed -i '$a\net.core.somaxconn = 262144' ${SYSCTL}
                        sed -i '$a\net.ipv4.tcp_max_orphans = 262144' ${SYSCTL}
                        sed -i '$a\net.ipv4.tcp_max_tw_buckets = 8192' ${SYSCTL}
                        sed -i '$a\net.ipv4.tcp_keepalive_probes = 5' ${SYSCTL}
                        sed -i '$a\net.core.netdev_max_backlog = 262144' ${SYSCTL}
                        sed -i '$a\vm.dirty_background_ratio = 5' ${SYSCTL}
                        sed -i '$a\vm.dirty_ratio = 10' ${SYSCTL}
                        sed -i '$a\vm.swappiness = 0' ${SYSCTL}
                        sed -i '$a\vm.vfs_cache_pressure = 62' ${SYSCTL}
                        sed -i '$a\vm.overcommit_memory = 1' ${SYSCTL}
                        echo "The Optimized end..."
                        exit 0
               else
                        echo "The '${SYSCTL}' file is not find"
                        exit 1
                fi
                
		;;

    [nN][oO]|[nN])
		echo "No"
       	;;

    *)
		echo "Invalid input..."
		exit 1
		;;
esac
