# import numpy

init = [
    150,  # reward amount
    (60 * 60) * 24 * 3,  # duration in epoch seconds
    200,  # start of stake
]

reward_proportionality = init[0] / init[1]


def payout(duration):
    return reward_proportionality * duration


print(payout(60 * 60 * 5))


word = "hellofren".encode("utf-8")
print(len(word))

