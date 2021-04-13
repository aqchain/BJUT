// SPDX-License-Identifier: GPL-3.0
pragma solidity 0.5.4;

contract Test {

    struct File{
        string fid;
        address owner;
    }

    mapping(string=>File)  files;

    function set(string memory fid,address owner) public{
        File storage file = files[fid];
        require(file.owner==address(0),"file exsited.");
        files[fid] = File(fid,owner);
    }

    function getOwner(string memory fid) public view returns(address){
        return files[fid].owner;
    }

    function test()public{

    }

}

