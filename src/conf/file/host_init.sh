#! /bin/bash

##################################################
#                                                #
#     Init script for host-base applications     #
#                                                #
#     Author: Feo Lee                            #
#     email: lih002@cmft.com                     #
#     Version: 0.1                               #
#     Date: 2020-01-02                           #
#                                                #
##################################################

readonly ERROR="\033[31mError:\033[0m"
readonly HINT="\033[34mHint:\033[0m"
readonly OK="\033[32mOK:\033[0m"

readonly OS_RELEASE="/etc/os-release"
readonly SYS_CONFIG_FILE="/etc/profile"

readonly PKG_URL_BASE="http://mirrors.cmftdc.cn/software/"
readonly JAVA_PKG_NAME="jdk-8u211-linux-x64.tar.gz"
readonly PY2_PKG_NAME="Python-2.7.10_el7.tar.gz"
readonly PY3_PKG_NAME="Python-3.6.5_el7.tar.gz"

readonly BASE_PATH="/app"
readonly JAVA8_HOME="${BASE_PATH}/jdk1.8.0_121"
readonly TOMCAT7_HOME="${BASE_PATH}/apache-tomcat-7.0.75"
readonly PY27_HOME="${BASE_PATH}/Python-2.7.10"
readonly PY36_HOME="${BASE_PATH}/Python-3.6.5"
readonly PIP_CONFIG_FILE="/etc/pip.conf"

# Print usage
print_usage() {
    printf "This script need 4 arguments.\n"
    printf "%-10s %-20s %-10s\n" args1: \<app_type\> eg.jar/py/ng...
    printf "%-10s %-20s %-10s\n" args2: \<unit_name\> eg.dcp-app/fhrms-web...
    printf "%-10s %-20s %-10s\n" args3: \<run_env\> eg.prd/uat/dev
    printf "%-10s %-20s %-10s\n" args4: \<net_zone\> eg.cmft-core/cmrh-dmz...
}

# OS check
check_os() {
    if [ ! -f ${OS_RELEASE} ]; then
        printf "${ERROR} ${OS_RELEASE} is missing, unsupported OS!\n"
        exit 1
    fi
    source ${OS_RELEASE}
    case ${ID} in
        centos|rhel)
        printf "${HINT} OS ID is ${ID}.\n"
        ;;
        *)
        printf "${HINT} OS ID is ${ID}.\n"
        printf "${ERROR} Unsupported OS!\n"
        exit 1
        ;;
    esac
}

# Init app dir
init_dir() {
    if [ -d ${BASE_PATH}/appsystems/${1}_${2}_${3}_${4} ]; then
        printf "${HINT} ${BASE_PATH}/appsystems/${1}_${2}_${3}_${4} already exist, skip.\n"
    else
        mkdir -p ${BASE_PATH}/appsystems/${1}_${2}_${3}_${4}/apps
        mkdir -p ${BASE_PATH}/appsystems/${1}_${2}_${3}_${4}/logs
    fi

    if [ -d ${BASE_PATH}/backup ]; then
        printf "${HINT} ${BASE_PATH}/backup already exist, skip.\n"
    else
        mkdir ${BASE_PATH}/backup
    fi

    if [ -d ${BASE_PATH}/logs/rtlog ]; then
        printf "${HINT} ${BASE_PATH}/logs/rtlog already exist, skip.\n"
    else
        mkdir -p ${BASE_PATH}/logs/rtlog
    fi

    if [ -L ${BASE_PATH}/logs/rtlog/${2} ]; then
        printf "${HINT} ${BASE_PATH}/logs/rtlog/${2} already exist, skip.\n"
    else
        ln -s ${BASE_PATH}/appsystems/${1}_${2}_${3}_${4}/logs ${BASE_PATH}/logs/rtlog/${2}
    fi

    chown -R mwop. ${BASE_PATH}/appsystems ${BASE_PATH}/backup ${BASE_PATH}/logs
    chmod -R 755 ${BASE_PATH}/appsystems ${BASE_PATH}/backup ${BASE_PATH}/logs
}

# Install JDK
check_jdk() {
    # printf "${HINT} checking java installation requirements...\n"
    if [ -d ${JAVA8_HOME} ]; then
        printf "${ERROR} ${JAVA8_HOME} already exist!\n"
        exit 1
    fi

    if [ -d ${TOMCAT7_HOME} ]; then
        printf "${ERROR} ${TOMCAT7_HOME} already exist!\n"
        exit 1
    fi

    java -version > /dev/null 2>&1
    if [ ${?} -eq 0 ]; then
        printf "${ERROR} java command already exist!\n"
        exit 1
    fi
}
install_jdk() {
    printf "${HINT} [java install - 1/5] download package\n"
    curl -o ${JAVA_PKG_NAME} ${PKG_URL_BASE}/${JAVA_PKG_NAME}

    printf "${HINT} [java install - 2/5] unzip package\n"
    tar -xzf ${JAVA_PKG_NAME} -C ${BASE_PATH}

    printf "${HINT} [java install - 3/5] install jdk&tomcat\n"
#    cp -a app/* ${BASE_PATH}
    chown -R mwop. ${JAVA8_HOME} ${TOMCAT7_HOME}
    # chmod -R 755 ${JAVA8_HOME} ${TOMCAT7_HOME}

    printf "${HINT} [java install - 4/5] config env.\n"
    echo -e "# Set for java" >> ${SYS_CONFIG_FILE}
    echo -e "export JAVA_HOME=${JAVA8_HOME}" >> ${SYS_CONFIG_FILE}
    echo -e "export PATH=\${JAVA_HOME}/bin:\${PATH}" >> ${SYS_CONFIG_FILE}
    echo -e "export CLASSPATH=.:\${JAVA_HOME}/lib/dt.jar:\${JAVA_HOME}/lib/tools.jar" >> ${SYS_CONFIG_FILE}

    printf "${HINT} [java install - 5/5] clean workspace\n"
    rm -rf app/ ${JAVA_PKG_NAME}
}

# Install Python
check_python() {
    if [ -d ${PY27_HOME} ]; then
        printf "${ERROR} ${PY27_HOME} already exist!\n"
        exit 1
    fi

    if [ -d ${PY36_HOME} ]; then
        printf "${ERROR} ${PY36_HOME} already exist!\n"
        exit 1
    fi

    python27 -V > /dev/null 2>&1
    if [ ${?} -eq 0 ]; then
        printf "${ERROR} command python27 already exist!\n"
        exit 1
    fi

    python36 -V > /dev/null 2>&1
    if [ ${?} -eq 0 ]; then
        printf "${ERROR} command python36 already exist!\n"
        exit 1
    fi
}
install_python() {
    cd ${BASE_PATH}
    printf "${HINT} [python install - 1/5] download package\n"
    curl -o ${PY2_PKG_NAME} ${PKG_URL_BASE}/${PY2_PKG_NAME}
    curl -o ${PY3_PKG_NAME} ${PKG_URL_BASE}/${PY3_PKG_NAME}

    printf "${HINT} [python install - 2/5] install package\n"
    tar -xzf ${PY2_PKG_NAME}
    tar -xzf ${PY3_PKG_NAME}
    chown -R mwop. ${PY27_HOME} ${PY36_HOME}
    chmod -R 755 ${PY27_HOME} ${PY36_HOME}

    printf "${HINT} [python install - 3/5] config pip env\n"
    if [ -f ${PIP_CONFIG_FILE} ]; then
        printf "${HINT} ${PIP_CONFIG_FILE} already exist, skip.\n"
    else
        printf "${HINT} ${PIP_CONFIG_FILE} is missing, creating...\n"
        touch ${PIP_CONFIG_FILE}
        chmod 644 ${PIP_CONFIG_FILE}
        printf "${OK} ${PIP_CONFIG_FILE} created.\n"
    fi
    echo -e "[global]" >> ${PIP_CONFIG_FILE}
    echo -e "index-url = http://mirrors.cmrh.com/pypi/web/simple" >> ${PIP_CONFIG_FILE}
    echo -e "trusted-host = mirrors.cmrh.com" >> ${PIP_CONFIG_FILE}

    printf "${HINT} [python install - 4/5] config sys env\n"
    echo -e "# Set for python" >> ${SYS_CONFIG_FILE}
    echo -e "export PATH=\${PATH}:${PY27_HOME}/sbin:${PY36_HOME}/sbin" >> ${SYS_CONFIG_FILE}
    echo -e "export PIP_CONFIG_FILE=/etc/pip.conf" >> ${SYS_CONFIG_FILE}

    printf "${HINT} [python install - 5/5] clean workspace\n"
    rm -f ${PY2_PKG_NAME} ${PY3_PKG_NAME}
}

# Main
if [ ${UID} -ne 0 ]; then
    printf "${ERROR} Please run as root!\n"
    exit 1
fi

check_os

if [ ${#} -ne 4 ]; then
    printf "${ERROR} Arguments error!\n"
    print_usage
    exit 1
fi


case ${1} in
    jar|tm)
    init_dir ${1} ${2} ${3} ${4}
    check_jdk
    install_jdk
    ;;
    py|py2|py3)
    init_dir ${1} ${2} ${3} ${4}
    check_python
    install_python
    ;;
    ng)
    init_dir ${1} ${2} ${3} ${4}
    ;;
    *)
    printf "${ERROR} Unsupported app type, please contact admin!\n"
    exit 1
    ;;
esac