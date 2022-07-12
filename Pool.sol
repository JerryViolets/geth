pragma solidity 0.5.16;

import ".././IERC20.sol";
import ".././ERC20.sol";
contract ABSwapPool is ERC20 {
    address public token0;
    address public token1;
    
    constructor(address _token0, address _token1) public{
        token0 = _token0;
        token1 = _token1;
    }
    uint public reserve0;
    uint public reserve1;
    uint public constant INITIAL_SUPPLY = 10**5;


    function add(uint amount0, uint amount1)public{
        assert(IERC20(token0).transferFrom(msg.sender,address(this),amount0));
        assert(IERC20(token1).transferFrom(msg.sender,address(this),amount1));
        if (reserve0 == 0 && reserve1 == 0){
            _mint(msg.sender, INITIAL_SUPPLY);
        }else{
            uint currentSupply;
            currentSupply = totalSupply;
        }
    }

    function operate() public{
        
    }
    function getAmountOut(uint amountIn,address fromToken) public returns(uint amountOut,uint newReserve0,uint newReserve1){
        uint k = reserve0 * reserve1;
        // uint newReserve0;
        // uint newReserve1;
        // uint amountOut;
        newReserve0 = amountIn + reserve0;
        newReserve1 = k / newReserve0;
        amountOut = reserve1 - newReserve1;
    }
    function swap(uint amountIn, address fromToken,address toToken,address to)public{
        (uint amountOut, uint newReserve0, uint newReserve1) = getAmountOut(amountIn, fromToken);
        assert(IERC20(fromToken).transferFrom(msg.sender,address(this),amountIn));
        assert(IERC20(toToken).transfer(to,amountOut));
        reserve0 = newReserve0;
        reserve1 = newReserve1;
    }
    function _mint(address user, uint amount) internal{

    }
}