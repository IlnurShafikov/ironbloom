# Game algorithm: Ways, 5 reels, base 2-4-4-4-2 / free spins 5x4, symbol 1 stacks (working title: Ironbloom)

Symbol IDs start at 1 (ID 0 is the reserved empty symbol and is not used in this model).
1-4 high-pay (1 is the most valuable and is the "herd" symbol that forms stacks on the strips), 5-9 low-pay, 10 Scatter (does not pay), 11 Wild x1 (base game only), 12-15 Wild x2 / x3 / x4 / x5 (free spins only). Wild IDs are contiguous (11-15) and the ID minus 10 equals the multiplier.

1. **Bet.** One spin costs `simple_cost_multiplier.csv` = 20 units. Paytable payouts are in the same units (1 unit = 1/20 of the total bet). Spin/round result in total bets = total round win / 20.

2. **Reel set selection.** One draw per spin, with no draw per reel.
   - Base game: a group is chosen by weights from `base/distributions.csv` (groups 1, 2, 3 with weights 30, 50, 20).
   - Each free spin: a group is chosen independently by weights from `free/distributions.csv` (groups 1, 2, 3, 4 with weights 4195, 3995, 1800, 10; group 4 is a rare tail set with heavier symbol 1 stacks and Wild x4/x5 cells).
   - All 5 reels use the strips of the chosen `group_N.csv` (column = reel 1..5, row = strip position).

3. **Reel stops.** For each reel independently, the stop position is chosen uniformly among all non-`-1` positions of that reel's column. Strips are circular (position 1 follows the last one). The visible window of reel _r_ is the `H_r` consecutive symbols at positions stop, stop+1, ..., stop+H_r-1 (mod strip length); the symbol at the stop position is row 0 (top). `H_r` comes from `base_symbols_on_reel.csv` (2,4,4,4,2 = 256 ways) in the base game and from `free_symbols_on_reel.csv` (4,4,4,4,4 = 1024 ways) in free spins. All strips in the delivered groups have equal length per phase (base 50, free 40), so no `-1` padding is used.

4. **Board modifications.** None. Stacks of symbol 1 are not a board modification: they exist only as blocks of 2-4 identical consecutive cells on the strips , so one reel window can show several cells of symbol 1. Wild multipliers are not drawn: they are defined by the symbol ID on the strip (`wild_multipliers.csv`).

5. **Win evaluation (ways, left to right).** Evaluated for each paying symbol S = 1..9 independently.
   - A reel _matches_ S if its window contains at least one cell that is S or a Wild (IDs in `wild_ids.csv`). Wild substitutes for symbols 1-9 only, never for the Scatter, and has no payout of its own. Wilds appear only on reels 2-4, so reel 1 must contain a real S.
   - Let k be the number of consecutive matching reels counted from reel 1 (k = 0 if reel 1 does not match). Matching stops at the first non-matching reel.
   - If k >= 1 and `paytable.csv[S][k]` > 0, symbol S pays `paytable[S][k] x W(S,k)`. Only this longest match counts for S; shorter counts of the same S do not pay additionally. A payout of 0 in the paytable means the count does not pay (symbol 1 pays from 2 reels, symbols 2-9 from 3 reels).
   - `W(S,k) = product over reels r = 1..k of m_r`, where `m_r` is the sum over all cells of the window of reel r that hold S or a Wild of that cell's multiplier: 1 for S, and `wild_multipliers.csv` value for a Wild (11 -> 1, 12 -> 2, 13 -> 3, 14 -> 4, 15 -> 5). This equals the sum over all individual ways (one cell per reel) of the product of the cell multipliers along the way. In the base game only Wild 10 is on the strips, so every multiplier is 1 and `m_r` is a cell count.
   - Scatter (ID `scatter_symbol_id` = 10) pays nothing. The delivered strips are built so that a window contains at most one Scatter per reel, so the Scatter count on the board is 0..5.

6. **Win modifiers.** The only modifiers are the Wild multipliers of step 5 (they multiply across reels inside a way and are summed inside a reel, never applied to the spin total). There is no other multiplier or doubling.

7. **Summation.** The spin win is the sum of the payouts of all symbols S = 1..9.

8. **Win cap.** `max_win_limit.csv` = 10000 total bets = 200000 units, applied to the whole round (trigger spin win plus all free spins), after Wild multipliers. When the cumulative round win reaches or exceeds 200000 units, it is truncated to exactly 200000, the round ends immediately, and unplayed free spins are forfeited. In the base game without a feature the same cap applies to the single spin win.

9. **Free spins.**
   - **Trigger:** if the base game board shows 3, 4 or 5 Scatters, the feature starts and awards 8, 12 or 16 free spins (`fs_init_counter.csv`). The win of the trigger spin (step 7) is paid and is added to the round total.
   - **Free spin:** each free spin uses the free spins window (5x4), a free group drawn per step 2, and steps 3-7. Its win is added to the round total. The base game groups are never used in free spins.
   - **Retrigger:** after the win of a free spin is evaluated, if its board shows 3, 4 or 5 Scatters, 8, 12 or 16 spins are added to the remaining counter (same table `fs_init_counter.csv`).
   - **End:** the session ends when the remaining counter is 0, or when the cap of step 8 is reached. There is no limit on the number of retriggers or on the total number of free spins.

10. **Spin result.** Round payout in total bets = round win (after the cap) / `simple_cost_multiplier.csv` = / 20.

## RTP versions

RTP versions 90, 92, 94 and 96 use the same algorithm, strips and paytable. They differ only in `distributions_base.csv`, `distributions_free.csv` and `exp_rtp.csv` (see the folders rtp_90 ... rtp_96).

## Notes for the implementer

- Wild 11 (x1) is on base strips only, Wilds 12-15 are on free strips only, both only on reels 2-4.
- The payouts of different symbols never overlap or exclude each other: the same cell may contribute to several symbols through a Wild.
- Strips contain no two identical adjacent symbols (circularly), except symbol 1, which appears in stacks of 2-4 identical cells. Scatters on a strip are at least 4 positions apart.
