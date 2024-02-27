RUN:
	
    You can run the source file from the folder using the following command:
    go run main.go <COMMAND> <FLAGS>
    Or you can run the executable from command line using the following syntax:
    DomainSlicer_V1.1 <COMMAND> <FLAGS>




INSTALLATION:
	
    
    git clone -u https://github.com/BubbaCodeX/DomainSlicerv1.1.git
COMMANDS :

    Parse Command
    Pings hosts from a list and sorts them based on the http responses,
	enabling a smoother bug bounty workflow

    Usage
    To use the Parse command:
    DomainSlicer_V1.1 Parse -f [path/to/hosts/file] -w [number_of_workers]
    -f, --filepath: Path to a file containing a list of hosts.
    -w, --workers: Amount of workers to use (default is 5).
    
    Example
     Parse -f /path/to/hosts.txt -w 10
    License
    This command is licensed under the same license as the project.



    
