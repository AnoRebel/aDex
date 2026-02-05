export namespace events {
	
	export class EventBus {
	
	
	    static createFrom(source: any = {}) {
	        return new EventBus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	
	    }
	}

}

export namespace exec {
	
	export class Cmd {
	    Path: string;
	    Args: string[];
	    Env: string[];
	    Dir: string;
	    Stdin: any;
	    Stdout: any;
	    Stderr: any;
	    ExtraFiles: os.File[];
	    SysProcAttr?: syscall.SysProcAttr;
	    Process?: os.Process;
	    // Go type: os
	    ProcessState?: any;
	    Err: any;
	    WaitDelay: number;
	
	    static createFrom(source: any = {}) {
	        return new Cmd(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Path = source["Path"];
	        this.Args = source["Args"];
	        this.Env = source["Env"];
	        this.Dir = source["Dir"];
	        this.Stdin = source["Stdin"];
	        this.Stdout = source["Stdout"];
	        this.Stderr = source["Stderr"];
	        this.ExtraFiles = this.convertValues(source["ExtraFiles"], os.File);
	        this.SysProcAttr = this.convertValues(source["SysProcAttr"], syscall.SysProcAttr);
	        this.Process = this.convertValues(source["Process"], os.Process);
	        this.ProcessState = this.convertValues(source["ProcessState"], null);
	        this.Err = source["Err"];
	        this.WaitDelay = source["WaitDelay"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace os {
	
	export class Process {
	    Pid: number;
	
	    static createFrom(source: any = {}) {
	        return new Process(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Pid = source["Pid"];
	    }
	}

}

export namespace syscall {
	
	export class Credential {
	    Uid: number;
	    Gid: number;
	    Groups: number[];
	    NoSetGroups: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Credential(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Uid = source["Uid"];
	        this.Gid = source["Gid"];
	        this.Groups = source["Groups"];
	        this.NoSetGroups = source["NoSetGroups"];
	    }
	}
	export class SysProcIDMap {
	    ContainerID: number;
	    HostID: number;
	    Size: number;
	
	    static createFrom(source: any = {}) {
	        return new SysProcIDMap(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ContainerID = source["ContainerID"];
	        this.HostID = source["HostID"];
	        this.Size = source["Size"];
	    }
	}
	export class SysProcAttr {
	    Chroot: string;
	    Credential?: Credential;
	    Ptrace: boolean;
	    Setsid: boolean;
	    Setpgid: boolean;
	    Setctty: boolean;
	    Noctty: boolean;
	    Ctty: number;
	    Foreground: boolean;
	    Pgid: number;
	    Pdeathsig: number;
	    Cloneflags: any;
	    Unshareflags: any;
	    UidMappings: SysProcIDMap[];
	    GidMappings: SysProcIDMap[];
	    GidMappingsEnableSetgroups: boolean;
	    AmbientCaps: any[];
	    UseCgroupFD: boolean;
	    CgroupFD: number;
	    PidFD?: number;
	
	    static createFrom(source: any = {}) {
	        return new SysProcAttr(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Chroot = source["Chroot"];
	        this.Credential = this.convertValues(source["Credential"], Credential);
	        this.Ptrace = source["Ptrace"];
	        this.Setsid = source["Setsid"];
	        this.Setpgid = source["Setpgid"];
	        this.Setctty = source["Setctty"];
	        this.Noctty = source["Noctty"];
	        this.Ctty = source["Ctty"];
	        this.Foreground = source["Foreground"];
	        this.Pgid = source["Pgid"];
	        this.Pdeathsig = source["Pdeathsig"];
	        this.Cloneflags = source["Cloneflags"];
	        this.Unshareflags = source["Unshareflags"];
	        this.UidMappings = this.convertValues(source["UidMappings"], SysProcIDMap);
	        this.GidMappings = this.convertValues(source["GidMappings"], SysProcIDMap);
	        this.GidMappingsEnableSetgroups = source["GidMappingsEnableSetgroups"];
	        this.AmbientCaps = source["AmbientCaps"];
	        this.UseCgroupFD = source["UseCgroupFD"];
	        this.CgroupFD = source["CgroupFD"];
	        this.PidFD = source["PidFD"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace terminal {
	
	export class Terminal {
	    ID: string;
	    Width: number;
	    Height: number;
	    Command?: exec.Cmd;
	    // Go type: os
	    PTY?: any;
	    WorkingDir: string;
	
	    static createFrom(source: any = {}) {
	        return new Terminal(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.Width = source["Width"];
	        this.Height = source["Height"];
	        this.Command = this.convertValues(source["Command"], exec.Cmd);
	        this.PTY = this.convertValues(source["PTY"], null);
	        this.WorkingDir = source["WorkingDir"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace utils {
	
	export class PlatformInfo {
	    OS: string;
	    Arch: string;
	    IsWindows: boolean;
	    IsLinux: boolean;
	    IsDarwin: boolean;
	    IsUnix: boolean;
	    HasPTY: boolean;
	    HasInotify: boolean;
	    HasAudioAPI: boolean;
	    PathSeparator: string;
	    LineEnding: string;
	    ExecutableExt: string;
	    SupportsSymlinks: boolean;
	
	    static createFrom(source: any = {}) {
	        return new PlatformInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.OS = source["OS"];
	        this.Arch = source["Arch"];
	        this.IsWindows = source["IsWindows"];
	        this.IsLinux = source["IsLinux"];
	        this.IsDarwin = source["IsDarwin"];
	        this.IsUnix = source["IsUnix"];
	        this.HasPTY = source["HasPTY"];
	        this.HasInotify = source["HasInotify"];
	        this.HasAudioAPI = source["HasAudioAPI"];
	        this.PathSeparator = source["PathSeparator"];
	        this.LineEnding = source["LineEnding"];
	        this.ExecutableExt = source["ExecutableExt"];
	        this.SupportsSymlinks = source["SupportsSymlinks"];
	    }
	}
	export class FeatureDetection {
	    Platform: PlatformInfo;
	    Features: Record<string, boolean>;
	
	    static createFrom(source: any = {}) {
	        return new FeatureDetection(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Platform = this.convertValues(source["Platform"], PlatformInfo);
	        this.Features = source["Features"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

