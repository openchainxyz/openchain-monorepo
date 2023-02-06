const examples = {
    raw: `function transfer(address,uint256)
function transferFrom(address,address,uint256)

event Transfer(address,address,uint256)

error InsufficientBalance(address)`,
    huff: `#define function transfer(address,uint256) nonpayable returns ()
#define function transferFrom(address,address,uint256) nonpayable returns ()

#define event Transfer(address,address,uint256)

#define error InsufficientBalance(address)`,
    abi: JSON.stringify([
        {
            constant: false,
            inputs: [
                {
                    name: '_from',
                    type: 'address',
                },
                {
                    name: '_to',
                    type: 'address',
                },
                {
                    name: '_value',
                    type: 'uint256',
                },
            ],
            name: 'transferFrom',
            outputs: [
                {
                    name: '',
                    type: 'bool',
                },
            ],
            payable: false,
            stateMutability: 'nonpayable',
            type: 'function',
        },
        {
            constant: false,
            inputs: [
                {
                    name: '_to',
                    type: 'address',
                },
                {
                    name: '_value',
                    type: 'uint256',
                },
            ],
            name: 'transfer',
            outputs: [
                {
                    name: '',
                    type: 'bool',
                },
            ],
            payable: false,
            stateMutability: 'nonpayable',
            type: 'function',
        },
        {
            anonymous: false,
            inputs: [
                {
                    indexed: true,
                    name: 'from',
                    type: 'address',
                },
                {
                    indexed: true,
                    name: 'to',
                    type: 'address',
                },
                {
                    indexed: false,
                    name: 'value',
                    type: 'uint256',
                },
            ],
            name: 'Transfer',
            type: 'event',
        },
        {
            inputs: [
                {
                    name: 'who',
                    type: 'address',
                },
            ],
            name: 'InsufficientBalance',
            type: 'error',
        },
    ]),
    solidity: `contract ERC20 {
    event Transfer(address indexed from, address indexed to, uint amount);
    
    error InsufficientBalance(address who);

    function transfer(address to, uint amount) external {}
    function transferFrom(address from, address to, uint amount) external {}
}`,
    vyper: `event Transfer:
   sender: indexed(address)
   receiver: indexed(address)
   value: uint256
   
@external
def transfer(to: address, amount: uint256):
    pass
    
@external
def transferFrom(_from: address, to: address, amount: uint256):
    pass
`,
};

export default examples;
