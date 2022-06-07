package jenkins_xml

const vmTmplV1 = `<?xml version='1.0' encoding='UTF-8'?>
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
  
    <!-- 构建应用 -->
    {{ xml.build_shell }}

    <!-- 构建后命令，主要是shell脚本 -->
    {{ xml.after_build_shell }}

  </builders>
  <publishers/>
  <buildWrappers/>
</project>`
func GetVMJenkinsXml() string {
	return vmTmplV1
}

const WSTargetPath  = `target`
