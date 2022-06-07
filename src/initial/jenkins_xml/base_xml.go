package jenkins_xml

import (
	"github.com/astaxie/beego"
	"strings"
)

const BASE_TMPL_V1 = `<?xml version='1.0' encoding='UTF-8'?>
<project>
  <actions/>
  <description/>
  <keepDependencies>false</keepDependencies>
  <properties/>
  <scm class="hudson.plugins.git.GitSCM" plugin="git@3.7.0">
    <configVersion>2</configVersion>
    <userRemoteConfigs>
      <hudson.plugins.git.UserRemoteConfig>
        <url>{{ xml.git_http_url }}</url>
        <credentialsId>git-devops-oper</credentialsId>
      </hudson.plugins.git.UserRemoteConfig>
    </userRemoteConfigs>
    <branches>
      <hudson.plugins.git.BranchSpec>
        <name>{{ xml.git_sha }}</name>
      </hudson.plugins.git.BranchSpec>
    </branches>
    <doGenerateSubmoduleConfigurations>false</doGenerateSubmoduleConfigurations>
    <submoduleCfg class="list"/>
    <extensions/>
  </scm>

  <scmCheckoutRetryCount>5</scmCheckoutRetryCount>
  <!-- 选择节点 -->
  {{ xml.jenkins_node }}
  <disabled>false</disabled>
  <blockBuildWhenDownstreamBuilding>false</blockBuildWhenDownstreamBuilding>
  <blockBuildWhenUpstreamBuilding>false</blockBuildWhenUpstreamBuilding>
  <jdk>jdk1.8.0_60</jdk>
  <triggers/>
  <concurrentBuild>false</concurrentBuild>
  <builders>

    <!-- 构建前命令，主要是shell脚本 -->
    {{ xml.before_build_shell }}
  
    <!-- dockfile镜像替换，配置初始化，主要是shell脚本，因部分前置命令可能会复制dockfile，放在前置命令后面 -->
    <hudson.tasks.Shell>
   		<command>{{ xml.replace_base_image }}</command>
    </hudson.tasks.Shell>

    <!-- 构建应用 -->
    {{ xml.build_shell }}

    <!-- 打包到dockfile -->
    <org.jenkinsci.plugins.conditionalbuildstep.singlestep.SingleConditionalBuilder plugin="conditional-buildstep@1.3.5">
      <condition class="org.jenkins_ci.plugins.run_condition.core.FileExistsCondition" plugin="run-condition@1.0">
        <file>Dockerfile</file>
        <baseDir class="org.jenkins_ci.plugins.run_condition.common.BaseDirectory$Workspace"/>
      </condition>
      <buildStep class="com.cloudbees.dockerpublish.DockerBuilder" plugin="docker-build-publish@1.3.2">
        <server plugin="docker-commons@1.6"/>
        <registry plugin="docker-commons@1.6">
          <url>{{ xml.harbor_url }}</url><credentialsId>{{ xml.harbor_credentials_id }}</credentialsId>
        </registry>
        <repoName>{{ xml.docker_repo_name }}</repoName>
        <noCache>false</noCache>
        <forcePull>true</forcePull>
        <dockerfilePath>Dockerfile</dockerfilePath>
        <skipBuild>false</skipBuild>
        <skipDecorate>false</skipDecorate>
        <repoTag>latest</repoTag>
        <skipPush>false</skipPush>
        <createFingerprint>true</createFingerprint>
        <skipTagLatest>true</skipTagLatest>
        <forceTag>true</forceTag>
      </buildStep>
      <runner class="org.jenkins_ci.plugins.run_condition.BuildStepRunner$Fail" plugin="run-condition@1.0"/>
    </org.jenkinsci.plugins.conditionalbuildstep.singlestep.SingleConditionalBuilder>

    <!-- 构建后命令，主要是shell脚本 -->
    {{ xml.after_build_shell }}

  </builders>
  <publishers/>
  <buildWrappers/>
</project>`


// di/st/prd不同的替换命令
const DI_REPLACE_IMAGE_SHELL = `
sed -i "s#registry.cmrh.com:5000#harbor.uat.cmft.com/base#g" Dockerfile
sed -i "s#harbor.uat.cmft.com/base/library#harbor.uat.cmft.com/base#g" Dockerfile
docker logout harbor.uat.cmft.com`

const PRD_REPLACE_IMAGE_SHELL = `
sed -i "s#registry.cmrh.com:5000#harbor.cmft.com/base#g" Dockerfile
sed -i "s#harbor.cmft.com/base/library#harbor.cmft.com/base#g" Dockerfile
sed -i "s#harbor.uat.cmft.com#harbor.cmft.com#g" Dockerfile
docker logout harbor.cmft.com`

const DR_REPLACE_IMAGE_SHELL = `
sed -i "s#registry.cmrh.com:5000#harbor.cmft.com/base#g" Dockerfile
sed -i "s#harbor.cmft.com/base/library#harbor.cmft.com/base#g" Dockerfile
sed -i "s#harbor.uat.cmft.com#harbor.cmft.com#g" Dockerfile
docker logout harbor-dr.cmft.com`


// di/st/prd/dr 不同的harbor镜像和密钥
const DI_HARBOR_URL = "https://harbor.uat.cmft.com"
const DI_CREDENT_ID = "08607a9a-7c59-4f7e-81ab-87aa8f6d1cf0"

const PRD_HARBOR_URL = "https://harbor.cmft.com"
const PRD_CREDENT_ID = "e1fb94a6-6811-4123-a8a6-a0da24ee3dd3"

const DR_HARBOR_URL = "https://harbor-dr.cmft.com"
const DR_CREDENT_ID = "24b4bd9f-50f5-4bf2-8419-ad56a988a8d4"


// 节点选择，后续节点管单独的项目
const JENKINS_NODE = `<assignedNode>{{ var.jenkins_node_tag }}</assignedNode>
<canRoam>false</canRoam>`



func GetBaseJenkinsXml() string {
	now_runmode := beego.AppConfig.String("runmode")
	if now_runmode == "prd" {
		xml := strings.Replace(BASE_TMPL_V1, "{{ xml.replace_base_image }}", PRD_REPLACE_IMAGE_SHELL, -1)
		xml = strings.Replace(xml, "{{ xml.harbor_url }}", PRD_HARBOR_URL, -1)
		xml = strings.Replace(xml, "{{ xml.harbor_credentials_id }}", PRD_CREDENT_ID, -1)
		return xml
	} else if now_runmode == "dr" {
		// 从 prd-harbor拉取基础镜像，构建，将生成的镜像推送到harbor-dr
		xml := strings.Replace(BASE_TMPL_V1, "{{ xml.replace_base_image }}", DR_REPLACE_IMAGE_SHELL, -1)
		xml = strings.Replace(xml, "{{ xml.harbor_url }}", DR_HARBOR_URL, -1)
		xml = strings.Replace(xml, "{{ xml.harbor_credentials_id }}", DR_CREDENT_ID, -1)
		return xml
	} else {
		// 不是开发环境，用测试环境配置
		xml := strings.Replace(BASE_TMPL_V1, "{{ xml.replace_base_image }}", DI_REPLACE_IMAGE_SHELL, -1)
		xml = strings.Replace(xml, "{{ xml.harbor_url }}", DI_HARBOR_URL, -1)
		xml = strings.Replace(xml, "{{ xml.harbor_credentials_id }}", DI_CREDENT_ID, -1)
		return xml
	}
}

func GetBaseJenkinsXmlAssignNode(node_tag, xml string) string {
	assign_node_xml := "<canRoam>true</canRoam>"
	if strings.Trim(node_tag, " ") != "" {
		assign_node_xml= strings.Replace(JENKINS_NODE, "{{ var.jenkins_node_tag }}", node_tag, -1)
	}
	return strings.Replace(xml, "{{ xml.jenkins_node }}", assign_node_xml, -1)
}

// 通用构建，包括前置命令、构建命令和后置命令，都是通过shell实现。
const BASE_COMMON_SHELL = `
    <hudson.tasks.Shell>
      <command>{{ common_build_shell }}</command>
    </hudson.tasks.Shell>`
