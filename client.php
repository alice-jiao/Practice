<?php

error_reporting(E_ALL);
set_time_limit(0);
date_default_timezone_set( 'Asia/Chongqing');  
define('WORK_PATH', dirname(__FILE__));

echo "<h2>TCP/IP Connection</h2>\n";

const PORT = 1993;
const IP = "127.0.0.1";    
    
class Client{
    private $urlFilePath;
    private $logFilePath;
    private $backupFilePath;
    private $socket;

    public function __construct(){

        $this->urlFilePath = WORK_PATH. "/data/";

        $this->logFilePath = WORK_PATH . '/logs/client.log';

        $this->backupFilePath = WORK_PATH . '/backup/';

        $dirName = dirname($this->logFilePath);
        if (!is_dir($dirName)) {
            $this->addLog("creating log file dir : $dirName ...");
            if(!mkdir($dirName, 0777, true)) {
                $this->addLog("Unable to create log file dir : $dirName , exiting...");
            }
        }

        /*
         +-------------------------------
         *    @socket连接整个过程
         +-------------------------------
         *    @socket_create
         *    @socket_connect
         *    @socket_write
         *    @socket_read
         *    @socket_close
         +--------------------------------
         */
        $this->socket = socket_create(AF_INET, SOCK_STREAM, SOL_TCP);
        if ($this->socket < 0) {
            $this->addLog("socket_create() failed. reason: " . socket_strerror($this->socket));
        }else {
            $this->addLog("socket create() OK!");
        }

        $this->addlog("tring to connect " . IP . " port : " .PORT);
        $result = socket_connect($this->socket, IP, PORT);
        if ($result < 0) {
            $this->addLog("socket_connect() failed. reason: ".socket_strerror($result));
        }else {
            $this->addLog("connenct OK!");
        }
    }

    public function run(){
        $ret = array();
        while(true){
            
            $inArr = $this->get_files_by_ext($this->urlFilePath, 'url');
            if(!empty($inArr)){
                sleep(2);
                foreach ($inArr as $in) {
                    if(!in_array($in, $ret)){
                        $ret[] = $in;
                        exec("cp {$this->urlFilePath}{$in} {$this->backupFilePath}");
                        $fileContent = file_get_contents($this->urlFilePath.$in);
                        if(socket_write($this->socket, $fileContent, strlen($fileContent))) {
                            $this->addLog("发送到服务器信息成功！发送的内容为: {$fileContent}");
                            sleep(2);
                        }else{
                            $this->addLog("发送失败！ reason :". socket_strerror($this->socket));
                        }
                    }
                    

                }
            }else{
                $this->addLog("无URL文件"); 
                sleep(1);
            }
        }

    }


    /**
     * 增加一行日志,自动添加换行符
     * @param str $log 日志内容
     * @return
     */
    protected function addLog($log) {
        $log = date("Y-m-d H:i:s") . " {$log} \n";
        file_put_contents($this->logFilePath, $log, FILE_APPEND | LOCK_EX);
        echo date("Y-m-d H:i:s") . " {$log} \n";

    }



    protected function get_files_by_ext($path, $ext){
     
        $files = array();
     
        if (is_dir($path)){
            $handle = opendir($path); 

            while ($file = readdir($handle)) { 
                if ($file[0] == '.'){ 
                    continue; 
                }
                if (is_file($path.$file) && preg_match('/\.'.$ext.'$/', $file)){ 
                        $files[] = $file;
                } 
            }
            closedir($handle);
            sort($files);
        }
        return $files;
     
    }
    
}


$obj = new Client();
$obj->run();

?>