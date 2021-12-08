package main

import (
	
	"runtime"
	"os"
	"os/exec"
	"syscall"
	"io/ioutil"
	"path/filepath"
	"github.com/stonelike/Xyz/gorecurcopy"
	
	
)



func init(){
 //ここでexecを行う、なおos.Execだとさらに子プロセスができてしまう模様
  if os.Args[0] == "callInit"{
	  runtime.GOMAXPROCS(1)
	  runtime.LockOSThread()

	  prepare()

	  must(syscall.Exec(os.Args[1],os.Args[1:],os.Environ()))

	  panic("this is unreachable!")
  }
}

func cleanTemp(tempfileName string){
	os.RemoveAll(tempfileName)
}

func initialCopy(){
	cur,_ :=os.Getwd()
	
	
	for _, v:= range []string{"/bin","/lib","/lib64"}{
		path:=filepath.Join(cur,v)
		if _,err := os.Stat(path);err!=nil{
			//pathがなかったら新たに作る
			
			must(os.MkdirAll(path,0777))
		}
		gorecurcopy.CopyDirectory(v,path)
	}
}

func main(){
	cur,_ := os.Getwd()
	tempname,_ := ioutil.TempDir(cur,"tmp")
	os.Chdir(tempname)
	initialCopy()

    must(run(os.Args[1:]))
	
	//ここ動的にtempfileNameを評価するか、prepare時に確実にこのプログラムが終わったら
	//終わらせる保証をしたい
	defer cleanTemp(tempname)
}

func run(args []string) error {
	if len(args) == 0{
		args = []string{os.Getenv("SHELL")}
	}
	
	//ここでfork
	cmd := exec.Command("/proc/self/exe",args...)
	cmd.Args[0] = "callInit"
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags:syscall.CLONE_NEWUTS | syscall.CLONE_NEWUSER | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS,
		GidMappings:[]syscall.SysProcIDMap{
			{
				ContainerID:0,
				HostID:os.Getgid(),
				Size:1,
			},
		},
		UidMappings:[]syscall.SysProcIDMap{
			{
				ContainerID:0,
				HostID:os.Geteuid(),
				Size:1,
			},
		},
	}

	return cmd.Run()
}

func must(err error){
	if err != nil{
		panic(err)
	}
}

func prepare() {
	root,_:=os.Getwd()
	
	//後々マウントするやつを全部まとめる
	for _, v := range []string{"/proc","/etc"}{
	   path:=filepath.Join(root,v)
	   //pathがちゃんとあるか？
	   if _,err := os.Stat(path);err!=nil{
		   //pathがなかったら新たに作る
		   //mkdirだと/tmpしか作れない、pathに/tmp/procとか一つ以上を指定したいとき
		   //mkdirAll
		   must(os.MkdirAll(path,0700))
	   }
	}
	//chrootするときはchroot先に行ってしまうのでそこにprocがないとmountできない

	must(syscall.Mount("proc",filepath.Join(root,"/proc"),"proc",0,""))

	for _, v:= range []string{"/etc/resolv.conf","/etc/hosts"}{
		d,err := ioutil.ReadFile(v)
		must(err)
		must(ioutil.WriteFile(filepath.Join(root,v),d,0700))
	}
    

	must(syscall.Chroot(root))
	must(syscall.Chdir("/"))
  
	
}