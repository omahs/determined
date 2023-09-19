# Adapted from https://github.com/Lightning-AI/lightning/blob/master/examples/pytorch/basics/backbone_image_classifier.py.
# Copyright and license reproduced below.
#
# Copyright The Lightning AI team.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

"""MNIST backbone image classifier example.

To run: python backbone_image_classifier.py --trainer.max_epochs=50

"""
import logging
import os
from typing import Optional

import filelock
import numpy as np
import pytorch_lightning as pl
import torch
from pytorch_lightning import LightningDataModule, LightningModule
from pytorch_lightning.utilities.imports import _TORCHVISION_AVAILABLE
from torch.nn import functional as F
from torch.utils.data import DataLoader, Dataset, random_split
from torchvision import datasets

if _TORCHVISION_AVAILABLE:
    from torchvision import transforms

DATASETS_PATH = "/tmp"


class Backbone(torch.nn.Module):
    """
    >>> Backbone()  # doctest: +ELLIPSIS +NORMALIZE_WHITESPACE
    Backbone(
      (l1): Linear(...)
      (l2): Linear(...)
    )
    """

    def __init__(self, hidden_dim=128):
        super().__init__()
        self.l1 = torch.nn.Linear(28 * 28, hidden_dim)
        self.l2 = torch.nn.Linear(hidden_dim, 10)

    def forward(self, x):
        x = x.view(x.size(0), -1)
        x = torch.relu(self.l1(x))
        return torch.relu(self.l2(x))


class LitClassifier(LightningModule):
    """
    >>> LitClassifier(Backbone())  # doctest: +ELLIPSIS +NORMALIZE_WHITESPACE
    LitClassifier(
      (backbone): ...
    )
    """

    def __init__(self, backbone: Optional[Backbone] = None, lr: float = 0.0001):
        super().__init__()
        self.save_hyperparameters(ignore=["backbone"])
        if backbone is None:
            backbone = Backbone()
        self.backbone = backbone

    def forward(self, x):
        # use forward for inference/predictions
        return self.backbone(x)

    def training_step(self, batch, batch_idx):
        x, y = batch
        y_hat = self(x)
        loss = F.cross_entropy(y_hat, y)
        self.log("train_loss", loss, on_epoch=True)
        return loss

    def validation_step(self, batch, batch_idx):
        x, y = batch
        y_hat = self(x)
        loss = F.cross_entropy(y_hat, y)
        self.log("valid_loss", loss, on_step=True)

    def test_step(self, batch, batch_idx):
        x, y = batch
        y_hat = self(x)
        loss = F.cross_entropy(y_hat, y)
        self.log("test_loss", loss)

    def predict_step(self, batch, batch_idx, dataloader_idx=None):
        x, y = batch
        return self(x)

    def configure_optimizers(self):
        # self.hparams available because we called self.save_hyperparameters()
        return torch.optim.Adam(self.trainer.model.parameters(), lr=self.hparams.lr)


class FakeMNIST(Dataset):
    def __init__(self, num_samples=60000, image_size=28):
        """
        Args:
            num_samples (int): Number of fake samples to generate.
            image_size (int): Size of the image (MNIST is 28x28).
        """
        self.num_samples = num_samples
        self.image_size = image_size
        self.data = self._generate_fake_data()

    def _generate_fake_data(self):
        # Generate random grayscale images
        return np.random.randint(
            0, 256, (self.num_samples, 1, self.image_size, self.image_size)
        ).astype(np.uint8)

    def __len__(self):
        return self.num_samples

    def __getitem__(self, idx):
        # Since this is fake data, there's no real label. We'll just assign a random label.
        label = np.random.randint(0, 10)
        return torch.tensor(self.data[idx], dtype=torch.float32), label


class MyDataModule(LightningDataModule):
    def __init__(self, batch_size: int = 32, num_workers=4):
        super().__init__()
        with filelock.FileLock(os.path.join(os.getcwd(), "lock")):
            dataset = datasets.MNIST(
                "../data", train=True, download=True, transform=transforms.ToTensor()
            )
            self.mnist_test = datasets.MNIST(
                "../data", train=False, transform=transforms.ToTensor()
            )
        self.mnist_train, self.mnist_val = random_split(dataset, [55000, 5000])
        # self.mnist_train = FakeMNIST(5000)
        # self.mnist_val = FakeMNIST(5000)
        # self.mnist_test = FakeMNIST(5000)
        self.batch_size = batch_size
        self.num_workers = num_workers

    def train_dataloader(self):
        return DataLoader(
            self.mnist_train, batch_size=self.batch_size, num_workers=self.num_workers
        )

    def val_dataloader(self):
        return DataLoader(self.mnist_val, batch_size=self.batch_size, num_workers=self.num_workers)

    def test_dataloader(self):
        return DataLoader(self.mnist_test, batch_size=self.batch_size, num_workers=self.num_workers)

    def predict_dataloader(self):
        return DataLoader(self.mnist_test, batch_size=self.batch_size, num_workers=self.num_workers)


if __name__ == "__main__":
    from integration import build_determined_trainer, determined_core_init, get_hyperparameters

    import determined as det

    logging.basicConfig(level=logging.INFO, format=det.LOG_FORMAT)
    pl.seed_everything(7)
    hparams = get_hyperparameters()
    with determined_core_init() as core_context:
        trainer, resume_path = build_determined_trainer(
            core_context,
            logger=pl.loggers.CSVLogger(save_dir="/tmp/"),
            callbacks=[
                pl.callbacks.LearningRateMonitor(logging_interval="step"),
                pl.callbacks.progress.TQDMProgressBar(refresh_rate=10),
            ],
        )
        model = LitClassifier(lr=hparams.lr)
        data_module = MyDataModule()
        trainer.fit(model, data_module, ckpt_path=resume_path)
        trainer.test(model, datamodule=data_module)
