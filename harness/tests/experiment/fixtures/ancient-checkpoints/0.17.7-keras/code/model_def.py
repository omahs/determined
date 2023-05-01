import numpy as np
from tensorflow.keras import layers, losses, models, optimizers, utils

from determined.keras import TFKerasTrial


class MySequence(utils.Sequence):
    def __init__(self, batch_size, length, data, label):
        self.batch_size = batch_size
        self.length = length
        self.data = data
        self.label = label

    def __len__(self):
        return self.length

    def __getitem__(self, idx):
        return np.array([self.data] * self.batch_size), np.array([self.label] * self.batch_size)


class OneVarTrial(TFKerasTrial):
    """
    Models a simple one variable(y = wx) neural network, and a MSE loss function.
    """

    def __init__(self, context) -> None:
        self.context = context
        self.my_batch_size = self.context.get_per_slot_batch_size()  # type: int
        self.my_lr = self.context.get_hparams()["learning_rate"]
        self.context.env.container_gpus = []

    def build_model(self) -> models.Sequential:
        model = models.Sequential()
        model.add(
            layers.Dense(
                1, activation=None, use_bias=False, kernel_initializer="zeros", input_shape=(1,)
            )
        )
        model = self.context.wrap_model(model)

        optimizer = optimizers.SGD(learning_rate=self.my_lr)
        optimizer = self.context.wrap_optimizer(optimizer)

        model.compile(optimizer, losses.mean_squared_error, metrics=["accuracy"])

        return model

    def build_training_data_loader(self):
        return MySequence(self.context.get_per_slot_batch_size(), 10000, 1.0, 2.0)

    def build_validation_data_loader(self):
        return MySequence(self.context.get_per_slot_batch_size(), 100, 1.0, 2.0)
