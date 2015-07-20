Summary: vali agent for ops.
Name: vali-agent
Version: 1.0.2
Release: 141215000.el6
Requires: redhat-lsb-core >= 4.0-7
License: GPL
Group: nosa 
Source: vali-agent.tar.gz
URL: http://www.nosa.me
Packager: nosa 
BuildRoot: %{_tmppath}/%{name}-%{version}-root
AutoReqProv: no

%define userpath /home/op

%define debug_package %{nil}

%description
vali-agent: gangr agent.

%prep
%setup -c

%build
    #make %{?_smp_mflags}

%install
    [ ${RPM_BUILD_ROOT} != "/" ] && rm -rf ${RPM_BUILD_ROOT}
    install -d ${RPM_BUILD_ROOT}%{userpath}
    %{__cp} -r %{_builddir}/%{name}-%{version}  ${RPM_BUILD_ROOT}%{userpath}/vali-agent


%files
%defattr(-,root,root,755)
%{userpath}/vali-agent

%pre
    if [ -f "/etc/init.d/vali-agent" ];then
            service vali-agent stop
    fi

%post
    mv /home/op/vali-agent/conf/vali-agent.conf /etc/init.d/vali-agent 2>/dev/null
    rm -r /home/op/vali-agent/conf
    chmod +x /etc/init.d/vali-agent
    chkconfig --level 345 vali-agent on
    service vali-agent start
    exit 0

%preun
    if [ "$1" = "0" ];then  #包卸载文件删除之前的动作处理
        :
    fi
    if [ "$1" = "1" ];then  #包被更新时文件未卸载之前的动作处理
        :
    fi

%postun
    if [ "$1" = 0 ];then
      if [ -f "/etc/init.d/vali-agent" ];then
              service vali-agent stop
              chkconfig --del vali-agent
              rm /etc/init.d/vali-agent
      fi
    fi
    echo

%clean
  [ "$RPM_BUILD_ROOT" != "/" ] && rm -rf $RPM_BUILD_ROOT
