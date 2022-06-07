package jenkins

import "testing"

func TestCreateJob(t *testing.T) {
	var jenk_opr JenkOpr
	jenk_opr.BaseUrl = "http://jenkins-di1.sit.cmrh.com:8080/"
	jenk_opr.JobName = "cmft-di-uad-mdp-web-20191017"
	jenk_opr.ConfigXml = getSampleXml()
	tag := jenk_opr.CreateJob()
	t.Log("CreateJob:",tag)
}




func getSampleXml() string {
	return `<?xml version='1.1' encoding='UTF-8'?>
<project>
  <actions/>
  <description>部署专用</description>
  <keepDependencies>false</keepDependencies>
  <properties>
    <com.dabsquared.gitlabjenkins.connection.GitLabConnectionProperty plugin="gitlab-plugin@1.5.13">
      <gitLabConnection>Gitlab Jakes Connection</gitLabConnection>
    </com.dabsquared.gitlabjenkins.connection.GitLabConnectionProperty>
  </properties>
  <scm class="hudson.plugins.git.GitSCM" plugin="git@3.12.1">
    <configVersion>2</configVersion>
    <userRemoteConfigs>
      <hudson.plugins.git.UserRemoteConfig>
        <url>http://git.dev.cmrh.com/opstool/multi-deploy-web.git</url>
        <credentialsId>git-devops-oper</credentialsId>
      </hudson.plugins.git.UserRemoteConfig>
    </userRemoteConfigs>
    <branches>
      <hudson.plugins.git.BranchSpec>
        <name>master</name>
      </hudson.plugins.git.BranchSpec>
    </branches>
    <doGenerateSubmoduleConfigurations>false</doGenerateSubmoduleConfigurations>
    <submoduleCfg class="list"/>
    <extensions/>
  </scm>
  <scmCheckoutRetryCount>5</scmCheckoutRetryCount>
  <assignedNode>jenkins-1</assignedNode>
  <canRoam>false</canRoam>
  <disabled>false</disabled>
  <blockBuildWhenDownstreamBuilding>false</blockBuildWhenDownstreamBuilding>
  <blockBuildWhenUpstreamBuilding>false</blockBuildWhenUpstreamBuilding>
  <jdk>jdk1.8.0_60</jdk>
  <triggers/>
  <concurrentBuild>false</concurrentBuild>
  <builders>
    <hudson.tasks.Shell>
      <command>cnpm install
npm run build
        </command>
    </hudson.tasks.Shell>
    <hudson.tasks.Shell>
      <command>tar -zvcf dist.tar.gz dist</command>
    </hudson.tasks.Shell>
  </builders>
  <publishers/>
  <buildWrappers/>
</project>`
}