# chainswaps
Calculate number of blocks for CLTV in cross-chain swaps

## Examples

Here is from ETH to BTC case. ETH has block interval 12s. Bitcoin has 10 minutes.

Check if 1000 ETH blocks and 50 Bitcoin blocks are good:

```
$ chainswaps-simulate -in-interval 12s -in-blocks 1000 -out-blocks 50 -out-interval 10m
Incoming blockchain mines a block in 12s.
Outgoing blockchain mines a block in 10m0s.
Simulation for 1000 blocks (-5 blocks reserved for confirmation) of incoming blockchain, 50 blocks of outgoing blockchain, time reserve 1m0s...
In simulation the share of being out of time is 0.999999999947.
Too high pvalue! BAD!
```

No, it is not good. There is high probability that HTLC of 50 Bitcoin blocks
expires AFTER the HTLC of 1000 ETH blocks.

Let's try 10k ETH blocks against 50 Bitcoin blocks:
(warning: long running simulation)

```
$ chainswaps-simulate -in-interval 12s -in-blocks 10000 -out-blocks 50 -out-interval 10m
Incoming blockchain mines a block in 12s.
Outgoing blockchain mines a block in 10m0s.
Simulation for 10000 blocks (-5 blocks reserved for confirmation) of incoming blockchain, 50 blocks of outgoing blockchain, time reserve 1m0s...
In simulation the share of being out of time is 0.
GOOD pvalue from simulation!
```

Ok, let's now find max safe number of Bitcoin blocks to fit into 2000 ETH blocks:

```
$ chainswaps-simulate -in-interval 12s -in-blocks 2000 -out-interval 10m -calibrate
Incoming blockchain mines a block in 12s.
Outgoing blockchain mines a block in 10m0s.
Increased time reserve by 20m0s to 21m0s to calibrate.
Calibration for pvalue 1e-09...
Restored time reserve to 1m0s after calibration.
The number of blocks in outgoing blockchain: 10.
Simulation for 2000 blocks (-5 blocks reserved for confirmation) of incoming blockchain, 10 blocks of outgoing blockchain, time reserve 1m0s...
In simulation the share of being out of time is 0.
GOOD pvalue from simulation!
```

So maximum safe number of Bitcoin blocks is only 10 to safely fit into 2000 ETH blocks.

## Options

```
$ chainswaps-simulate -h
Usage of chainswaps-simulate:
  -calibrate
        Calibrate max out-blocks to make pvalue <= target pvalue
  -calibration-leeway duration
        Additional time to add to time reserve during calibration (default 20m0s)
  -in-blocks int
        Number of blocks to be mined in input blockchain (known exactly from incoming invoice) (default 100)
  -in-blocks-reserve int
        Reserve number in input blockchain to let force close tx to confirm (ok to set higher) (default 5)
  -in-fixed
        Input blockchain produces blocks with fixed intervals
  -in-interval duration
        Average delay between blocks in input blockchain (ok to set lower than it is) (default 2m30s)
  -out-blocks int
        Max number of blocks to be mined in output blockchain (will be set when paying the outgoing invoice) (default 42)
  -out-fixed
        Output blockchain produces blocks with fixed intervals
  -out-interval duration
        Average delay between blocks in output blockchain (ok to set higher than it is) (default 10m0s)
  -p-value float
        Target probability of not having enough time to collect the money from incoming channel (must be near zero) (default 1e-09)
  -time-reserve duration
        Leeway to have time to settle incoming invoice after learning knowing preimage from outgoing payment (ok to set higher) (default 1m0s)
  -trials int
        Number of trials in each blockchain simulation (higher - more precise) (default 1000000)
```

Use `-in-fixed` or `-out-fixed` in case you use a blockchain with fixed block intervals (e.g. Liquid).
