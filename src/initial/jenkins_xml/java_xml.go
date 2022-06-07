package jenkins_xml

// java项目模板类，主要是针对特定项目，将jenkins插件可视化
// java目前有三个插件应用非常广泛，有：maven, ant, gradle

const JAVA_ANT_BUILD_SHELL = `
    <hudson.tasks.Ant plugin="ant@1.4">
		<targets>copyForDocker</targets>
		<antName>apache-ant-1.9.6</antName>
		<buildFile>build.xml</buildFile>
	</hudson.tasks.Ant>`

const JAVA_MAVEN_BUILD_SHELL = `
    <hudson.tasks.Maven>
    	<targets>clean package</targets>
    	<mavenName>Maven 3.3.9</mavenName>
    	<pom>pom.xml</pom>
    	<usePrivateRepository>false</usePrivateRepository>
    	<settings class="jenkins.mvn.DefaultSettingsProvider" />
    	<globalSettings class="jenkins.mvn.DefaultGlobalSettingsProvider" />
    </hudson.tasks.Maven>`

const JAVA_GRADLE_BUILD_SHELL = `
    <hudson.tasks.Maven>
    	<targets>clean package</targets>
    	<mavenName>Maven 3.3.9</mavenName>
    	<pom>pom.xml</pom>
    	<usePrivateRepository>false</usePrivateRepository>
    	<settings class="jenkins.mvn.DefaultSettingsProvider" />
    	<globalSettings class="jenkins.mvn.DefaultGlobalSettingsProvider" />
    </hudson.tasks.Maven>`



