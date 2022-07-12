pragma solidity ^0.8.0;


library Math {
    /**
     * @dev Returns the addition of two unsigned integers, reverting on
     * overflow.
     *
     * Counterpart to Solidity's `+` operator.
     *
     * Requirements:
     * - Addition cannot overflow.
     */
    function add(uint256 a, uint256 b) internal pure returns (uint256) {
        uint256 c = a + b;
        require(c >= a, "SafeMath: addition overflow");

        return c;
    }

    /**
     * @dev Returns the subtraction of two unsigned integers, reverting on
     * overflow (when the result is negative).
     *
     * Counterpart to Solidity's `-` operator.
     *
     * Requirements:
     * - Subtraction cannot overflow.
     */
    function sub(uint256 a, uint256 b) internal pure returns (uint256) {
        require(b <= a, "SafeMath: subtraction overflow");
        uint256 c = a - b;

        return c;
    }

    /**
     * @dev Returns the multiplication of two unsigned integers, reverting on
     * overflow.
     *
     * Counterpart to Solidity's `*` operator.
     *
     * Requirements:
     * - Multiplication cannot overflow.
     */
    function mul(uint256 a, uint256 b) internal pure returns (uint256) {
        // Gas optimization: this is cheaper than requiring 'a' not being zero, but the
        // benefit is lost if 'b' is also tested.
        // See: https://github.com/OpenZeppelin/openzeppelin-solidity/pull/522
        if (a == 0) {
            return 0;
        }

        uint256 c = a * b;
        require(c / a == b, "SafeMath: multiplication overflow");

        return c;
    }

    /**
     * @dev Returns the integer division of two unsigned integers. Reverts on
     * division by zero. The result is rounded towards zero.
     *
     * Counterpart to Solidity's `/` operator. Note: this function uses a
     * `revert` opcode (which leaves remaining gas untouched) while Solidity
     * uses an invalid opcode to revert (consuming all remaining gas).
     *
     * Requirements:
     * - The divisor cannot be zero.
     */
    function div(uint256 a, uint256 b) internal pure returns (uint256) {
        // Solidity only automatically asserts when dividing by 0
        require(b > 0, "SafeMath: division by zero");
        uint256 c = a / b;
        // assert(a == b * c + a % b); // There is no case in which this doesn't hold

        return c;
    }

    /**
     * @dev Returns the remainder of dividing two unsigned integers. (unsigned integer modulo),
     * Reverts when dividing by zero.
     *
     * Counterpart to Solidity's `%` operator. This function uses a `revert`
     * opcode (which leaves remaining gas untouched) while Solidity uses an
     * invalid opcode to revert (consuming all remaining gas).
     *
     * Requirements:
     * - The divisor cannot be zero.
     */
    function mod(uint256 a, uint256 b) internal pure returns (uint256) {
        require(b != 0, "SafeMath: modulo by zero");
        return a % b;
    }
    function min(uint256 a, uint256 b) internal pure returns (uint256) {
        return a < b ? a : b;
    }
}
//SPDX-License-Identifier: Unlicense


import "../ERC20/ERC20.sol";

import "../ERC20/IERC20.sol";

contract LooneySwapPool is ERC20 {
  address public token0;
  address public token1;

  // Reserve of token 0
  uint public reserve0;

  // Reserve of token 1
  uint public reserve1;

  uint public constant INITIAL_SUPPLY = 10**5;

  constructor(address _token0, address _token1) ERC20("LooneyLiquidityProvider", "LP") {
    token0 = _token0;
    token1 = _token1;
    //_balances[msg.sender] = 10000;
    _mint(msg.sender,1000);
  }

  /**
   * Adds liquidity to the pool.
   * 1. Transfer tokens to pool
   * 2. Emit LP tokens
   * 3. Update reserves
   */
  function add(uint amount0, uint amount1) public {
    assert(IERC20(token0).transferFrom(msg.sender, address(this), amount0));
    assert(IERC20(token1).transferFrom(msg.sender, address(this), amount1));

    uint reserve0After = reserve0 + amount0;
    uint reserve1After = reserve1 + amount1;

    if (reserve0 == 0 && reserve1 == 0) {
      _mint(msg.sender, INITIAL_SUPPLY);
    } else {
      uint currentSupply = totalSupply();
      uint newSupplyGivenReserve0Ratio = reserve0After * currentSupply / reserve0;
      uint newSupplyGivenReserve1Ratio = reserve1After * currentSupply / reserve1;
      uint newSupply = Math.min(newSupplyGivenReserve0Ratio, newSupplyGivenReserve1Ratio);
      _mint(msg.sender, newSupply - currentSupply);
    }

    reserve0 = reserve0After;
    reserve1 = reserve1After;
  }

  /**
   * Removes liquidity from the pool.
   * 1. Transfer LP tokens to pool
   * 2. Burn the LP tokens
   * 3. Update reserves
   */
  function remove(uint liquidity) public {
    assert(transfer(address(this), liquidity));

    uint currentSupply = totalSupply();
    uint amount0 = liquidity * reserve0 / currentSupply;
    uint amount1 = liquidity * reserve1 / currentSupply;

    _burn(address(this), liquidity);

    assert(IERC20(token0).transfer(msg.sender, amount0));
    assert(IERC20(token1).transfer(msg.sender, amount1));
    reserve0 = reserve0 - amount0;
    reserve1 = reserve1 - amount1;
  }

  /**
   * Uses x * y = k formula to calculate output amount.
   * 1. Calculate new reserve on both sides
   * 2. Derive output amount
   */
  function getAmountOut (uint amountIn, address fromToken) public view returns (uint amountOut, uint _reserve0, uint _reserve1) {
    uint newReserve0;
    uint newReserve1;
    uint k = reserve0 * reserve1;

    // x (reserve0) * y (reserve1) = k (constant)
    // (reserve0 + amountIn) * (reserve1 - amountOut) = k
    // (reserve1 - amountOut) = k / (reserve0 + amount)
    // newReserve1 = k / (newReserve0)
    // amountOut = newReserve1 - reserve1

    if (fromToken == token0) {
      newReserve0 = amountIn + reserve0;
      newReserve1 = k / newReserve0;
      amountOut = reserve1 - newReserve1;
    } else {
      newReserve1 = amountIn + reserve1;
      newReserve0 = k / newReserve1;
      amountOut = reserve0 - newReserve0;
    }

    _reserve0 = newReserve0;
    _reserve1 = newReserve1;
  }

  /**
   * Swap to a minimum of `minAmountOut`
   * 1. Calculate new reserve on both sides
   * 2. Derive output amount
   * 3. Check output against minimum requested
   * 4. Update reserves
   */
  function swap(uint amountIn, uint minAmountOut, address fromToken, address toToken, address to) public {
    require(amountIn > 0 && minAmountOut > 0, 'Amount invalid');
    require(fromToken == token0 || fromToken == token1, 'From token invalid');
    require(toToken == token0 || toToken == token1, 'To token invalid');
    require(fromToken != toToken, 'From and to tokens should not match');

    (uint amountOut, uint newReserve0, uint newReserve1) = getAmountOut(amountIn, fromToken);

    require(amountOut >= minAmountOut, 'Slipped... on a banana');

    assert(IERC20(fromToken).transferFrom(msg.sender, address(this), amountIn));
    assert(IERC20(toToken).transfer(to, amountOut));

    reserve0 = newReserve0;
    reserve1 = newReserve1;
  }
}